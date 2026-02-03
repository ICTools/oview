# 🔍 Guide de vérification : Claude Code utilise-t-il oview ?

## TL;DR - Test rapide

```bash
# 1. Vérifier la config
./verify_mcp.sh

# 2. Lancer Claude Code
claude

# 3. Demander à Claude
> Use project_info to show me this project's configuration
```

Si Claude retourne les infos du projet (embeddings, chunks, etc.), **ça marche !** ✅

---

## 🎯 Pourquoi vérifier ?

Claude Code peut accéder à votre code de **deux façons** :

### ❌ Méthode 1 : Lecture directe (par défaut)
```
Vous → Claude Code → Read tool → Fichiers
```
Claude lit directement vos fichiers (comme `cat`, `grep`)

**Problème** :
- Pas de recherche sémantique
- Pas de compréhension du contexte
- Recherche par mots-clés uniquement

### ✅ Méthode 2 : Via MCP + RAG (ce qu'on veut)
```
Vous → Claude Code → MCP → oview → PostgreSQL (embeddings)
```
Claude utilise votre index RAG pour chercher sémantiquement

**Avantage** :
- Recherche par sens, pas par mots
- Résultats triés par pertinence
- Beaucoup plus rapide sur gros projets

---

## 📊 Vérifications automatiques

### Script de vérification complet

```bash
./verify_mcp.sh
```

**Ce qu'il vérifie :**
- ✅ Configuration MCP (`~/.claude/mcp_servers.json`)
- ✅ Binary oview accessible
- ✅ Projet initialisé (`.oview/project.yaml`)
- ✅ Database avec chunks indexés
- ✅ Commande `search` fonctionnelle
- ✅ MCP server démarre correctement

**Output attendu :**
```
✅ PASS: MCP configuration found
✅ PASS: oview found at /usr/local/bin/oview
✅ PASS: Project initialized
✅ PASS: 181 chunks indexed
✅ PASS: Search functionality works
```

---

## 🧪 Tests manuels

### Test 1 : Demander explicitement l'outil MCP

```bash
claude
```

```
> Use the tool 'project_info'
```

**Résultat attendu :**
```json
{
  "project_id": "224c26...",
  "project_slug": "oview",
  "embeddings": {
    "provider": "ollama",
    "model": "nomic-embed-text",
    "dim": 768
  },
  "database": {
    "chunk_count": 181
  }
}
```

**Si ça marche :** Claude utilise le MCP ! ✅

**Si ça échoue :** Claude dira "I don't have access to that tool" ❌

---

### Test 2 : Recherche sémantique

```
> Use search to find authentication code
```

**Résultat attendu :**
```
I found 5 results related to authentication:

1. cmd/init.go (validateClaudeAPI) - 92% similarity
2. cmd/init.go (validateOpenAIEmbeddings) - 88% similarity
...
```

**Comment vérifier :**
- Claude mentionne "search" ou "similarity"
- Résultats triés par pertinence (%)
- Trouve du code même avec des synonymes

---

### Test 3 : Fichier caché (preuve définitive)

Ce test **prouve** que Claude utilise l'index et pas les fichiers.

**Étape 1 : Créer un fichier NON indexé**
```bash
echo "SECRET_MARKER_12345" > /tmp/hidden_test.txt
```

**Étape 2 : Demander à Claude**
```
> Search for SECRET_MARKER_12345
```

**Résultat attendu :**
- ✅ Claude utilise MCP search
- ✅ Ne trouve PAS le marqueur (pas indexé)
- ✅ Retourne 0 résultats

**Si Claude trouve le fichier :**
- ❌ Claude lit directement les fichiers
- ❌ MCP n'est pas utilisé
- 🔧 Vérifier la configuration MCP

**Étape 3 : Indexer et ré-essayer**
```bash
# Ajouter le fichier au projet
mv /tmp/hidden_test.txt ./test_marker.txt

# Ré-indexer
oview index

# Redemander à Claude
> Search for SECRET_MARKER_12345
```

**Maintenant Claude devrait le trouver !** ✅

---

### Test 4 : Fichier renommé

**Étape 1 : Renommer sans ré-indexer**
```bash
mv cmd/search.go cmd/search_RENAMED.go
```

**Étape 2 : Demander à Claude**
```
> Where is the search command implementation?
```

**Résultat attendu (MCP) :**
```
The search command is in cmd/search.go
```
(Ancien chemin indexé)

**Résultat si lecture directe :**
```
The search command is in cmd/search_RENAMED.go
```
(Nouveau chemin filesystem)

**Étape 3 : Remettre comme avant**
```bash
mv cmd/search_RENAMED.go cmd/search.go
```

---

## 📡 Monitoring en temps réel

### Méthode 1 : Logs MCP

**Terminal 1 : Démarrer MCP avec logs**
```bash
oview mcp 2>&1 | tee /tmp/oview_mcp.log
```

**Terminal 2 : Utiliser Claude Code**
```bash
claude
> Search for database connection
```

**Terminal 3 : Surveiller les logs**
```bash
tail -f /tmp/oview_mcp.log
```

**Output attendu :**
```json
{"level":"info","message":"Starting oview MCP server..."}
{"level":"info","message":"MCP request: tools/call"}
{"method":"search","query":"database connection"}
```

Si vous voyez ces messages → MCP fonctionne ! ✅

---

### Méthode 2 : Monitoring PostgreSQL

**Activer les logs SQL :**
```bash
docker exec oview-postgres psql -U postgres -c "ALTER SYSTEM SET log_statement = 'all';"
docker exec oview-postgres psql -U postgres -c "SELECT pg_reload_conf();"
```

**Surveiller les requêtes :**
```bash
docker logs -f oview-postgres 2>&1 | grep "SELECT"
```

**Quand Claude cherche via MCP, vous verrez :**
```sql
SELECT id, path, symbol, content,
       1 - (embedding <=> '[0.21,-0.48,...]'::vector) as similarity
FROM chunks
ORDER BY embedding <=> '[0.21,-0.48,...]'::vector
LIMIT 5
```

La présence de `embedding <=>` prouve que c'est une recherche vectorielle ! ✅

---

### Méthode 3 : strace (avancé)

**Surveiller les fichiers ouverts :**
```bash
strace -e openat -f claude 2>&1 | grep -E '\.(go|js|py|php)' | tee /tmp/claude_files.log
```

**Demander à Claude de chercher du code**

**Analyser les accès :**
```bash
grep -c "cmd/" /tmp/claude_files.log
grep -c "src/" /tmp/claude_files.log
```

**Résultat attendu (MCP) :**
- 0-2 fichiers ouverts (juste la config)
- Pas de lecture massive de fichiers source

**Résultat sans MCP :**
- 10+ fichiers ouverts
- Lectures directes dans cmd/, src/, etc.

---

## 🚨 Problèmes courants

### Problème 1 : Claude ne voit pas les outils MCP

**Symptôme :**
```
> Use search
Claude: I don't have access to a 'search' tool
```

**Solutions :**

1. **Vérifier la config MCP**
   ```bash
   cat ~/.claude/mcp_servers.json
   # Doit contenir "oview"
   ```

2. **Vérifier le chemin oview**
   ```bash
   which oview
   # Doit retourner un chemin valide
   ```

3. **Redémarrer Claude Code**
   - Fermer complètement
   - Relancer : `claude`

4. **Vérifier les logs Claude**
   ```bash
   cat ~/.claude/logs/mcp-*.log
   ```

---

### Problème 2 : MCP timeout

**Symptôme :**
```
Claude: Tool execution timed out
```

**Solutions :**

1. **Vérifier Ollama**
   ```bash
   ollama list  # Doit montrer nomic-embed-text
   curl http://localhost:11434/api/tags  # Doit répondre
   ```

2. **Vérifier PostgreSQL**
   ```bash
   docker ps | grep oview-postgres  # Doit être running
   ```

3. **Tester manuellement**
   ```bash
   oview search "test" --limit 1
   # Doit fonctionner rapidement
   ```

---

### Problème 3 : Résultats non pertinents

**Symptôme :**
```
Search for "authentication"
Returns: README.md sections with low relevance
```

**Solutions :**

1. **Ré-indexer**
   ```bash
   oview index
   ```

2. **Vérifier le nombre de chunks**
   ```bash
   docker exec oview-postgres psql -U oview -d oview_oview -c \
     "SELECT COUNT(*) FROM chunks;"
   ```

   Si < 50 chunks : Pas assez de contenu indexé

3. **Benchmarker la pertinence**
   ```bash
   ./oview benchmark --queries 10
   # Avg Top Result doit être > 60%
   ```

---

## 📈 Benchmark de performance

### Lancer un benchmark complet

```bash
./oview benchmark --queries 10 -o benchmark.json
```

**Métriques importantes :**

```json
{
  "avg_embedding_time_ms": 25,      // Génération embedding
  "avg_search_time_ms": 25,         // Recherche + embedding
  "min_search_time_ms": 23,         // Meilleur cas
  "max_search_time_ms": 28,         // Pire cas
  "throughput_queries_per_sec": 40, // Requêtes/seconde
  "avg_result_relevance": 0.62      // Pertinence (0-1)
}
```

**Interprétation :**

| Métrique | Excellent | Bon | Acceptable | Lent |
|----------|-----------|-----|------------|------|
| Search time | < 100ms | < 500ms | < 1s | > 1s |
| Throughput | > 20 q/s | > 10 q/s | > 5 q/s | < 5 q/s |
| Relevance | > 80% | > 60% | > 40% | < 40% |

**Votre benchmark (Ollama local) :**
- ✅ 25ms search → **Excellent**
- ✅ 40 q/s → **Excellent**
- ✅ 62% relevance → **Bon**

---

## ✅ Checklist finale

Avant d'utiliser avec Claude Code :

- [ ] MCP configuré (`~/.claude/mcp_servers.json`)
- [ ] oview accessible (`which oview`)
- [ ] Projet initialisé (`ls .oview/project.yaml`)
- [ ] Database avec chunks (`./verify_mcp.sh`)
- [ ] Search fonctionne (`oview search "test"`)
- [ ] Benchmark correct (`oview benchmark`)
- [ ] Claude voit les outils (demander `List tools`)

Une fois tout validé :

```bash
claude
> Use project_info
> Search for authentication
> Get context for cmd/init.go
```

**Tout fonctionne ? Félicitations ! 🎉**

Claude Code utilise maintenant votre RAG oview pour comprendre votre code !

---

## 📚 Ressources

- **Script auto** : `./verify_mcp.sh`
- **Benchmark** : `./oview benchmark --help`
- **Logs MCP** : `~/.claude/logs/`
- **Guide simple** : `docs/SIMPLE_EXPLANATION.md`
- **Guide technique** : `docs/HOW_IT_WORKS.md`
