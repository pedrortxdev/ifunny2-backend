# API de Posts - Documentação

## Autenticação
Todos os endpoints que requerem autenticação precisam dos seguintes headers:
- `User-ID`: ID do usuário obtido no login
- `Authorization`: Token obtido no login

# Seção 1: Usando curl

## Endpoints Públicos (curl)

### Criar Usuário
```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{
    "nome": "Usuário Teste",
    "email": "teste@exemplo.com",
    "senha": "123456"
  }' \
  http://192.168.15.2:8080/usuarios
```

### Login
```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{
    "email": "teste@exemplo.com",
    "senha": "123456"
  }' \
  http://192.168.15.2:8080/login
```

### Listar Posts
```bash
curl -X GET http://192.168.15.2:8080/posts
```

## Endpoints Autenticados (curl)

### Criar Post
```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -H "User-ID: {USER_ID}" \
  -H "Authorization: {TOKEN}" \
  -d '{
    "nome": "Meu primeiro post",
    "img": "https://exemplo.com/imagem.jpg",
    "desc": "Esta é a descrição do meu post"
  }' \
  http://192.168.15.2:8080/posts
```

### Adicionar Comentário
```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -H "User-ID: {USER_ID}" \
  -H "Authorization: {TOKEN}" \
  -d '{
    "texto": "Meu comentário no post"
  }' \
  "http://192.168.15.2:8080/posts?comment=true&post_id={POST_ID}"
```

### Dar/Remover Like
```bash
curl -X PUT \
  -H "Content-Type: application/json" \
  -H "User-ID: {USER_ID}" \
  -H "Authorization: {TOKEN}" \
  "http://192.168.15.2:8080/posts?id={POST_ID}"
```

# Seção 2: Usando Bruno

## Endpoints Públicos (Bruno)

### Login
```bruno
meta {
  name: Login
  type: http
}

post {
  url: http://192.168.15.2:8080/login
  body: json
}

headers {
  Content-Type: application/json
}

body: json {
  {
    "email": "teste@exemplo.com",
    "senha": "123456"
  }
}
```

### Criar Usuário
```bruno
meta {
  name: Create User
  type: http
}

post {
  url: http://192.168.15.2:8080/usuarios
  body: json
}

headers {
  Content-Type: application/json
}

body: json {
  {
    "nome": "Usuário Teste",
    "email": "teste@exemplo.com",
    "senha": "123456"
  }
}
```

### Listar Posts
```bruno
meta {
  name: List Posts
  type: http
}

get {
  url: http://192.168.15.2:8080/posts
}
```

## Endpoints Autenticados (Bruno)

### Criar Post
```bruno
meta {
  name: Create Post
  type: http
}

post {
  url: http://192.168.15.2:8080/posts
  body: json
}

headers {
  Content-Type: application/json
  User-ID: {USER_ID}
  Authorization: {TOKEN}
}

body: json {
  {
    "nome": "Meu primeiro post",
    "img": "https://exemplo.com/imagem.jpg",
    "desc": "Esta é a descrição do meu post"
  }
}
```

### Adicionar Comentário
```bruno
meta {
  name: Add Comment
  type: http
}

post {
  url: http://192.168.15.2:8080/posts?comment=true&post_id={POST_ID}
  body: json
}

headers {
  Content-Type: application/json
  User-ID: {USER_ID}
  Authorization: {TOKEN}
}

body: json {
  {
    "texto": "Meu comentário no post"
  }
}
```

### Dar/Remover Like
```bruno
meta {
  name: Toggle Like
  type: http
}

put {
  url: http://192.168.15.2:8080/posts?id={POST_ID}
  body: none
}

headers {
  Content-Type: application/json
  User-ID: {USER_ID}
  Authorization: {TOKEN}
}
```

# Observações Gerais

1. Substitua `{USER_ID}` pelo ID retornado no login
2. Substitua `{TOKEN}` pelo token retornado no login
3. Substitua `{POST_ID}` pelo ID do post desejado

## Funcionalidades
1. O endpoint de like funciona como toggle:
   - Primeira chamada: adiciona o like
   - Segunda chamada: remove o like
   - E assim por diante...
2. Ao listar posts, os comentários são incluídos automaticamente na resposta

## Permissões
1. Endpoints públicos (não requerem autenticação):
   - Criar usuário
   - Login
   - Listar posts
2. Endpoints autenticados (requerem User-ID e Token):
   - Criar posts
   - Adicionar comentários
   - Dar/remover likes

# Troubleshooting

## Erros Comuns

### 1. Erro 400 - Bad Request com CORS
Se você receber um erro como:
```
POST http://192.168.15.2:8080/posts
400 - Bad Request
access-control-allow-headers: Content-Type, User-ID
```

Possíveis soluções:
1. Verifique se todos os headers necessários estão sendo enviados corretamente:
   - `Content-Type: application/json`
   - `User-ID: {seu_user_id}`
   - `Authorization: {seu_token}`
2. Certifique-se de que o corpo da requisição está no formato JSON correto
3. Verifique se o token de autorização está válido e não expirou

### 2. Formato das Requisições
- Todos os endpoints autenticados DEVEM incluir os headers `User-ID` e `Authorization`
- O header `Content-Type: application/json` é obrigatório para todas as requisições que enviam dados no corpo (POST/PUT)
- Os valores dos headers devem ser enviados sem as chaves `{}` 

### 3. Dicas de Depuração
- Ao receber erros de autenticação, tente fazer login novamente para obter um novo token
- Para endpoints autenticados, sempre verifique se está enviando tanto o `User-ID` quanto o `Authorization`
- Se estiver usando Bruno, verifique se substituiu todas as variáveis entre chaves `{}` pelos valores reais
