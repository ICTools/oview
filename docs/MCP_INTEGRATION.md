# 🔌 Intégration MCP avec Claude Code

Ce guide explique comment connecter oview avec Claude Code via le Model Context Protocol (MCP).

## 📋 Prérequis

1. **oview installé et configuré**
   ```bash
   oview install          # Infrastructure globale
   cd /path/to/project
   oview init            # Initialisation du projet
   oview up              # Base de données
   oview index           # Indexation du code
   ```

2. **Claude Code installé**
   - Suivez les instructions sur [claude.ai/code](https://claude.ai/code)

## 🚀 Configuration

### 1. Compiler oview avec le support MCP

```bash
cd /home/david/Documents/oview
go build -o oview .
sudo cp oview /usr/local/bin/oview
```

### 2. Configurer Claude Code

Ajoutez le serveur MCP à la configuration de Claude Code:

**Fichier:** `~/.claude/mcp_servers.json`

```json
{
  "oview": {
    "command": "oview",
    "args": ["mcp"],
    "description": "oview RAG system for semantic code search",
    "autoApprove": ["search", "get_context", "project_info"]
  }
}
```

**Note:** Remplacez le chemin si oview n'est pas dans `/usr/local/bin/`

### 3. Vérifier la configuration

```bash
# Test manuel du MCP server
cd /path/to/your/project
oview mcp

# Le serveur attend des commandes JSON-RPC sur stdin
# Ctrl+C pour quitter
```

### 4. Redémarrer Claude Code

Redémarrez Claude Code pour qu'il charge le nouveau serveur MCP.

## 🎯 Utilisation

Une fois configuré, Claude Code aura accès à trois outils:

### 1. **search** - Recherche sémantique

Claude peut rechercher dans votre codebase:

```
Utilisateur: "Où est implémentée l'authentification ?"

Claude: [utilise search("authentication logic")]
```

### 2. **get_context** - Contexte d'un fichier

Claude peut obtenir du contexte avant de modifier du code:

```
Utilisateur: "Modifie src/Controller/UserController.php"

Claude: [utilise get_context("src/Controller/UserController.php")]
        [comprend le contexte]
        [propose les modifications]
```

### 3. **project_info** - Informations du projet

Claude peut voir la configuration du projet:

```
Utilisateur: "Quel est le stack de ce projet ?"

Claude: [utilise project_info()]
        Ce projet utilise...
```

## 📊 Exemple de session

```
Utilisateur: Comment fonctionne le système de cache dans ce projet ?

Claude Code:
1. [Utilise search("cache system implementation")]
2. Trouve 5 chunks pertinents dans:
   - src/Service/CacheManager.php
   - config/packages/cache.yaml
   - src/EventListener/CacheListener.php
3. Analyse le code
4. Explique le système de cache

"Le système de cache utilise Redis avec une stratégie de TTL..."
```

## 🔧 Dépannage

### Le serveur MCP ne démarre pas

```bash
# Vérifier que oview est accessible
which oview

# Vérifier la configuration du projet
cd /path/to/project
cat .oview/project.yaml

# Tester manuellement
oview mcp
```

### Claude Code ne voit pas les outils

1. Vérifier `~/.claude/mcp_servers.json`
2. Redémarrer Claude Code
3. Vérifier les logs Claude Code

### Erreurs de connexion à la base de données

```bash
# Vérifier que PostgreSQL tourne
docker ps | grep oview-postgres

# Vérifier la base du projet
docker exec oview-postgres psql -U oview -l | grep oview_
```

### Les embeddings ne fonctionnent pas

**Pour Ollama:**
```bash
# Vérifier Ollama
ollama list

# Relancer si nécessaire
ollama serve &
```

**Pour OpenAI:**
```bash
# Vérifier la clé API
echo $OPENAI_API_KEY

# Ou dans .oview/project.yaml
grep api_key .oview/project.yaml
```

## 🎨 Personnalisation

### Changer le nombre de résultats par défaut

Dans votre workflow, vous pouvez demander à Claude:

```
"Recherche les 10 meilleurs exemples d'API REST dans le projet"
```

Claude utilisera automatiquement `search("API REST examples", limit=10)`

### Indexer plus de contexte

Modifiez `.oview/rag.yaml` pour inclure plus de fichiers:

```yaml
indexing:
  include_paths:
    - src/
    - config/
    - templates/
    - docs/          # Ajouter la documentation
    - scripts/       # Ajouter les scripts
```

Puis ré-indexez:
```bash
oview index
```

## 🔐 Sécurité

- **Données locales**: Tout reste sur votre machine
- **Pas de télémétrie**: oview ne communique pas avec des serveurs externes
- **Embeddings locaux**: Utilisez Ollama pour un système 100% local
- **Clés API**: Si vous utilisez OpenAI, les clés sont dans `~/.bashrc` ou `.oview/project.yaml` (ne pas commit!)

## 📈 Performance

### Taille de l'index

```bash
# Voir les statistiques
docker exec oview-postgres psql -U oview -d oview_yourproject -c "
  SELECT
    COUNT(*) as chunks,
    pg_size_pretty(pg_total_relation_size('chunks')) as size
  FROM chunks;
"
```

### Vitesse de recherche

- Recherche sémantique: ~50-200ms (dépend de l'index HNSW)
- Génération d'embeddings:
  - Ollama local: ~100-500ms
  - OpenAI API: ~200-1000ms

## 🚀 Prochaines étapes

1. **Auto-refresh**: Ré-indexer automatiquement après un commit
2. **Cache embeddings**: Mettre en cache les requêtes fréquentes
3. **Multi-projets**: Rechercher dans plusieurs projets simultanément
4. **Filtres avancés**: Filtrer par langage, type, date, etc.

## 📚 Ressources

- [Documentation MCP](https://modelcontextprotocol.io/)
- [Claude Code](https://claude.ai/code)
- [pgvector](https://github.com/pgvector/pgvector)
- [Ollama](https://ollama.ai/)

## 💬 Support

Problème ou question ? Ouvrez une issue sur GitHub!
