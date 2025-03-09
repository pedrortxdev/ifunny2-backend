package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go-dev/config"
	"go-dev/models"
	"log"
	"net/http"
	"strconv"
	"time"
)

var db *sql.DB

func generateToken(userID int, email string) string {
	data := fmt.Sprintf("%d:%s:%d", userID, email, time.Now().Unix())
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}

func checkAuth(r *http.Request) (int, error) {
	userID := r.Header.Get("User-ID")
	token := r.Header.Get("Authorization")

	if userID == "" || token == "" {
		return 0, fmt.Errorf("não autenticado")
	}

	uid, err := strconv.Atoi(userID)
	if err != nil {
		return 0, fmt.Errorf("ID de usuário inválido")
	}

	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM usuarios WHERE id = ?)", uid).Scan(&exists)
	if err != nil {
		return 0, err
	}
	if !exists {
		return 0, fmt.Errorf("usuário não encontrado")
	}

	return uid, nil
}

func main() {
	fmt.Println("Servidor iniciando...")

	var err error
	db, err = config.InitDB()
	if err != nil {
		log.Fatalf("Erro ao conectar ao banco de dados: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS usuarios (
			id INT AUTO_INCREMENT PRIMARY KEY,
			nome VARCHAR(100) NOT NULL,
			email VARCHAR(100) NOT NULL UNIQUE,
			senha VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		log.Fatalf("Erro ao criar tabela usuarios: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS posts (
			id INT AUTO_INCREMENT PRIMARY KEY,
			nome VARCHAR(255) NOT NULL,
			img VARCHAR(255),
			descricao TEXT,
			likes INT DEFAULT 0,
			user_post INT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_post) REFERENCES usuarios(id)
		);
	`)
	if err != nil {
		log.Fatalf("Erro ao criar tabela posts: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS post_likes (
			post_id INT NOT NULL,
			user_id INT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (post_id, user_id),
			FOREIGN KEY (post_id) REFERENCES posts(id),
			FOREIGN KEY (user_id) REFERENCES usuarios(id)
		);
	`)
	if err != nil {
		log.Fatalf("Erro ao criar tabela post_likes: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS comentarios (
			id INT AUTO_INCREMENT PRIMARY KEY,
			post_id INT NOT NULL,
			user_id INT NOT NULL,
			texto TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (post_id) REFERENCES posts(id),
			FOREIGN KEY (user_id) REFERENCES usuarios(id)
		);
	`)
	if err != nil {
		log.Fatalf("Erro ao criar tabela comentarios: %v", err)
	}

	http.HandleFunc("/usuarios", handleCreateUser)
	http.HandleFunc("/login", handleLogin)
	http.HandleFunc("/posts", handlePosts)

	fmt.Println("Servidor escutando em [::]:8080 (IPv6) e :8080 (IPv4)")
	log.Fatal(http.ListenAndServe("[::]:8080", nil))
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	var login models.Login
	if err := json.NewDecoder(r.Body).Decode(&login); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var user models.Usuario
	err := db.QueryRow("SELECT id, nome, email, senha FROM usuarios WHERE email = ?", login.Email).
		Scan(&user.ID, &user.Nome, &user.Email, &user.Senha)

	if err == sql.ErrNoRows {
		http.Error(w, "Usuário não encontrado", http.StatusUnauthorized)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if user.Senha != login.Senha {
		http.Error(w, "Senha incorreta", http.StatusUnauthorized)
		return
	}

	token := generateToken(user.ID, user.Email)

	response := models.LoginResponse{
		ID:    user.ID,
		Nome:  user.Nome,
		Email: user.Email,
		Token: token,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleCreateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	var user models.Usuario
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := db.Exec(
		"INSERT INTO usuarios (nome, email, senha) VALUES (?, ?, ?)",
		user.Nome, user.Email, user.Senha,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()
	user.ID = int(id)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func handlePosts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, User-ID, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	switch r.Method {
	case http.MethodPost:
		if r.URL.Query().Get("comment") != "" {
			addComment(w, r)
		} else {
			createPost(w, r)
		}
	case http.MethodGet:
		listPosts(w, r)
	case http.MethodPut:
		likePost(w, r)
	default:
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
	}
}

func createPost(w http.ResponseWriter, r *http.Request) {
	uid, err := checkAuth(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var post models.Post
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	post.UserID = uid

	result, err := db.Exec(
		"INSERT INTO posts (nome, img, descricao, user_post) VALUES (?, ?, ?, ?)",
		post.Nome, post.Imagem, post.Descricao, post.UserID,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()
	post.ID = int(id)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(post)
}

func listPosts(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`
		SELECT 
			p.id, p.nome, p.img, p.descricao, p.likes, p.user_post, 
			DATE_FORMAT(p.created_at, '%Y-%m-%d %H:%i:%s') as created_at,
			u.nome as usuario_nome
		FROM posts p
		JOIN usuarios u ON p.user_post = u.id
		ORDER BY p.created_at DESC
	`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		var usuarioNome string
		var imgNull, descNull sql.NullString
		var createdAtStr string
		err := rows.Scan(
			&post.ID, &post.Nome, &imgNull, &descNull, &post.Likes,
			&post.UserID, &createdAtStr, &usuarioNome,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		post.CreatedAt, err = time.Parse("2006-01-02 15:04:05", createdAtStr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if imgNull.Valid {
			post.Imagem = &imgNull.String
		}
		if descNull.Valid {
			post.Descricao = &descNull.String
		}

		comentarios, err := getComentarios(post.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		post.Comentarios = comentarios

		posts = append(posts, post)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

func getComentarios(postID int) ([]models.Comentario, error) {
	rows, err := db.Query(`
		SELECT 
			c.id, c.post_id, c.user_id, c.texto, 
			DATE_FORMAT(c.created_at, '%Y-%m-%d %H:%i:%s') as created_at
		FROM comentarios c
		WHERE c.post_id = ?
		ORDER BY c.created_at DESC
	`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comentarios []models.Comentario
	for rows.Next() {
		var comentario models.Comentario
		var createdAtStr string
		err := rows.Scan(
			&comentario.ID, &comentario.PostID, &comentario.UserID,
			&comentario.Texto, &createdAtStr,
		)
		if err != nil {
			return nil, err
		}

		comentario.CreatedAt, err = time.Parse("2006-01-02 15:04:05", createdAtStr)
		if err != nil {
			return nil, err
		}

		comentarios = append(comentarios, comentario)
	}
	return comentarios, nil
}

func likePost(w http.ResponseWriter, r *http.Request) {
	uid, err := checkAuth(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	postID := r.URL.Query().Get("id")
	if postID == "" {
		http.Error(w, "ID do post não fornecido", http.StatusBadRequest)
		return
	}

	pid, err := strconv.Atoi(postID)
	if err != nil {
		http.Error(w, "ID do post inválido", http.StatusBadRequest)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	var exists bool
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM post_likes WHERE post_id = ? AND user_id = ?)", pid, uid).Scan(&exists)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if exists {
		_, err = tx.Exec("DELETE FROM post_likes WHERE post_id = ? AND user_id = ?", pid, uid)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = tx.Exec("UPDATE posts SET likes = likes - 1 WHERE id = ?", pid)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		_, err = tx.Exec("INSERT INTO post_likes (post_id, user_id) VALUES (?, ?)", pid, uid)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = tx.Exec("UPDATE posts SET likes = likes + 1 WHERE id = ?", pid)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	err = tx.Commit()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var post models.Post
	var imgNull, descNull sql.NullString
	var createdAtStr string
	err = db.QueryRow(`
		SELECT 
			p.id, p.nome, p.img, p.descricao, p.likes, p.user_post,
			DATE_FORMAT(p.created_at, '%Y-%m-%d %H:%i:%s') as created_at
		FROM posts p
		WHERE p.id = ?
	`, pid).Scan(
		&post.ID, &post.Nome, &imgNull, &descNull, &post.Likes,
		&post.UserID, &createdAtStr,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	post.CreatedAt, err = time.Parse("2006-01-02 15:04:05", createdAtStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if imgNull.Valid {
		post.Imagem = &imgNull.String
	}
	if descNull.Valid {
		post.Descricao = &descNull.String
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(post)
}

func addComment(w http.ResponseWriter, r *http.Request) {
	uid, err := checkAuth(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	postID := r.URL.Query().Get("post_id")
	if postID == "" {
		http.Error(w, "ID do post não fornecido", http.StatusBadRequest)
		return
	}

	pid, err := strconv.Atoi(postID)
	if err != nil {
		http.Error(w, "ID do post inválido", http.StatusBadRequest)
		return
	}

	var comment models.Comentario
	if err := json.NewDecoder(r.Body).Decode(&comment); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := db.Exec(
		"INSERT INTO comentarios (post_id, user_id, texto) VALUES (?, ?, ?)",
		pid, uid, comment.Texto,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()
	comment.ID = int(id)
	comment.PostID = pid
	comment.UserID = uid
	comment.CreatedAt = time.Now()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(comment)
}
