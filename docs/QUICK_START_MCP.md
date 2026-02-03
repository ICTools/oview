# 🚀 Quick Start: Connecter Claude Code à oview

## En 3 minutes chrono ⏱️

### 1. Vérifier que oview est indexé

```bash
cd /path/to/your/project

# Vérifier l'indexation
docker exec oview-postgres psql -U oview -d oview_oview -c "SELECT COUNT(*) FROM chunks;"
```

Vous devriez voir un nombre > 0. Sinon:
```bash
oview index
```

### 2. Configurer Claude Code

Créez ou éditez `~/.claude/mcp_servers.json`:

```bash
mkdir -p ~/.claude
cat > ~/.claude/mcp_servers.json << 'JSON'
{
  "mcpServers": {
    "oview": {
      "command": "oview",
      "args": ["mcp"]
    }
  }
}
JSON
```

### 3. Tester le MCP server

```bash
# Test rapide
cd /path/to/your/project
oview mcp &
MCP_PID=$!

# Envoyer une requête de test (JSON-RPC)
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' | nc localhost -

# Tuer le process
kill $MCP_PID
```

### 4. Utiliser dans Claude Code

Lancez Claude Code dans votre projet:

```bash
cd /path/to/your/project
claude
```

Puis testez:

```
> Recherche les fonctions qui gèrent l'authentification
```

Claude utilisera automatiquement l'outil `search` du MCP server oview ! 🎉

## 🔍 Exemples d'utilisation

**Recherche sémantique:**
```
> Où est le code qui gère les erreurs 404 ?
> Comment fonctionne le système de cache ?
> Trouve les tests pour la classe UserService
```

**Contexte avant modification:**
```
> Je veux modifier src/Controller/HomeController.php, 
  peux-tu me donner le contexte ?
```

**Informations du projet:**
```
> Quel est le stack technique de ce projet ?
> Combien de chunks sont indexés ?
```

## ✅ Vérification

Pour vérifier que tout fonctionne:

1. Ouvrir Claude Code dans votre projet
2. Taper: "Utilise l'outil project_info"
3. Claude devrait retourner les infos du projet

Si ça ne marche pas, voir `docs/MCP_INTEGRATION.md` pour le dépannage.

## 🎯 C'est tout !

Claude Code est maintenant connecté à votre RAG oview et peut:
- ✅ Rechercher sémantiquement dans votre code
- ✅ Obtenir du contexte pertinent automatiquement  
- ✅ Comprendre l'architecture de votre projet

Enjoy! 🚀
