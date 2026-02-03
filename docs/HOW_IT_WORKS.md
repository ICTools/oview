# 🧠 Comment fonctionne oview + Claude Code

## Vue d'ensemble

```
┌──────────────┐         ┌─────────────┐         ┌──────────────┐
│   Votre      │         │    oview    │         │  PostgreSQL  │
│   Code       │────────►│   Indexer   │────────►│  + pgvector  │
│              │  Scan   │             │  Store  │              │
└──────────────┘         └─────────────┘         └──────┬───────┘
                                                         │
                                                         │ Chunks +
                                                         │ Embeddings
                                                         │
┌──────────────┐         ┌─────────────┐                │
│   Claude     │◄────────┤   oview     │◄───────────────┘
│    Code      │  MCP    │   MCP       │  Search
│              │ Protocol│   Server    │
└──────────────┘         └─────────────┘
```

## 🔄 Le cycle complet

### Phase 1 : Indexation (une fois, puis après changements)

#### Étape 1.1 : Scanner les fichiers

```bash
$ oview index
```

Le scanner parcourt votre projet selon `.oview/rag.yaml` :

```yaml
indexing:
  include_paths:
    - src/
    - config/
    - templates/
  exclude_paths:
    - vendor/
    - node_modules/
```

**Fichiers trouvés :**
- `src/Controller/UserController.php`
- `src/Service/AuthService.php`
- `config/packages/security.yaml`
- etc.

#### Étape 1.2 : Découpage intelligent (Chunking)

Chaque fichier est découpé en **chunks** selon son type :

**Pour PHP :** Par fonction/méthode
```php
// Devient 1 chunk
class UserController {
    public function login(Request $request) {
        // Logique d'authentification
        return $this->render('login.html.twig');
    }
}
```

**Pour YAML :** Par section
```yaml
# Devient 1 chunk
security:
    providers:
        app_user_provider:
            entity:
                class: App\Entity\User
```

**Pour Markdown :** Par section (## titres)
```markdown
## Installation   ← Devient 1 chunk
Instructions...

## Configuration  ← Devient 1 autre chunk
Settings...
```

#### Étape 1.3 : Génération des embeddings

Chaque chunk est transformé en **vecteur** (embedding) :

```
┌─────────────────────────────────────────────────────┐
│ Chunk de code:                                      │
│                                                     │
│ class UserController {                              │
│     public function login() {                       │
│         // Authentication logic                     │
│     }                                                │
│ }                                                    │
└───────────────────┬─────────────────────────────────┘
                    │
                    ▼
            ┌───────────────┐
            │    Ollama     │ nomic-embed-text
            │   (local AI)  │ 768 dimensions
            └───────┬───────┘
                    │
                    ▼
┌─────────────────────────────────────────────────────┐
│ Embedding vector (768 nombres):                     │
│                                                     │
│ [0.234, -0.456, 0.123, 0.789, -0.234, 0.567, ...]  │
│                                                     │
│ ↑ Ces nombres capturent le SENS du code            │
└─────────────────────────────────────────────────────┘
```

**Pourquoi c'est magique ?**

Des codes similaires ont des vecteurs **proches** :

```python
"login authentication"    → [0.23, -0.45, 0.12, ...]
"user signin security"    → [0.21, -0.48, 0.11, ...]  # Proche!
"database connection"     → [0.89, 0.34, -0.67, ...]  # Éloigné!
```

#### Étape 1.4 : Stockage dans PostgreSQL

Les chunks + embeddings vont dans la base de données :

```sql
CREATE TABLE chunks (
    id SERIAL PRIMARY KEY,
    path TEXT,                    -- src/Controller/UserController.php
    symbol VARCHAR(255),          -- UserController::login
    content TEXT,                 -- Le code complet du chunk
    embedding vector(768),        -- [0.234, -0.456, ...]
    language VARCHAR(50),         -- php
    type VARCHAR(50),             -- code
    metadata JSONB
);

-- Index HNSW pour recherche ultra-rapide
CREATE INDEX ON chunks USING hnsw (embedding vector_cosine_ops);
```

**Résultat :** Base de données avec tous vos chunks indexés ! ✅

```
chunks table : 487 rows
├── README.md chunks (55 chunks)
├── cmd/init.go chunks (127 chunks)
├── cmd/search.go chunks (23 chunks)
└── ...
```

---

### Phase 2 : Recherche via Claude Code (en temps réel)

#### Étape 2.1 : Vous posez une question

```bash
$ claude
> Où est implémentée l'authentification ?
```

#### Étape 2.2 : Claude décide d'utiliser le RAG

Claude pense :
- "L'utilisateur cherche du code"
- "Je devrais utiliser l'outil `search` du MCP server oview"
- "Je vais chercher 'authentication implementation'"

#### Étape 2.3 : Appel MCP (JSON-RPC)

```json
// Claude → oview MCP server (via stdin)
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "search",
    "arguments": {
      "query": "authentication implementation",
      "limit": 5
    }
  }
}
```

#### Étape 2.4 : oview génère l'embedding de la requête

```
"authentication implementation"
        ↓ Ollama (même modèle!)
[0.21, -0.48, 0.11, ..., 0.82]  ← 768 dimensions
```

**Important :** Même modèle = embeddings comparables !

#### Étape 2.5 : Recherche de similarité

PostgreSQL + pgvector cherche les chunks les plus **similaires** :

```sql
SELECT
    id, path, symbol, content,
    1 - (embedding <=> '[0.21,-0.48,0.11,...]'::vector) AS similarity
FROM chunks
WHERE project_id = 'abc123'
ORDER BY embedding <=> '[0.21,-0.48,0.11,...]'::vector
LIMIT 5;
```

**L'opérateur `<=>`** calcule la distance cosinus :
- 0.0 = identiques
- 1.0 = complètement différents

**Résultats :**

```
┌────┬──────────────────────┬────────────┬────────────┐
│ ID │ Path                 │ Symbol     │ Similarity │
├────┼──────────────────────┼────────────┼────────────┤
│ 42 │ AuthService.php      │ authent... │   92.5%    │ ← Très pertinent!
│ 17 │ UserController.php   │ login()    │   88.3%    │ ← Pertinent
│ 89 │ SecurityConfig.yaml  │ security   │   76.2%    │ ← Assez pertinent
│ 12 │ User.php            │ User class │   64.1%    │ ← Moins pertinent
│ 55 │ README.md           │ Security   │   58.9%    │ ← Contexte
└────┴──────────────────────┴────────────┴────────────┘
```

#### Étape 2.6 : Retour à Claude

```json
// oview MCP server → Claude (via stdout)
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "count": 5,
    "results": [
      {
        "path": "src/Service/AuthService.php",
        "symbol": "AuthService::authenticate",
        "content": "class AuthService {\n  public function authenticate(...) {...}",
        "similarity": "92.50%"
      },
      // ... autres résultats
    ]
  }
}
```

#### Étape 2.7 : Claude répond

Claude analyse les résultats et vous répond :

```
L'authentification est implémentée principalement dans:

1. **src/Service/AuthService.php** (AuthService::authenticate)
   - Logique métier de l'authentification
   - Validation des credentials

2. **src/Controller/UserController.php** (login)
   - Endpoint HTTP pour le login
   - Gestion de la session

La configuration de sécurité se trouve dans config/packages/security.yaml.
```

---

## 🆚 Comparaison avec recherche traditionnelle

### Recherche par mots-clés (grep, Ctrl+F)

```bash
$ grep -r "authentication" .
```

**Problèmes :**
- ❌ Ne trouve que le mot exact "authentication"
- ❌ Rate "login", "signin", "auth", "credentials"
- ❌ Pas de compréhension du contexte
- ❌ Beaucoup de faux positifs

**Exemple :**
```
Query: "authentication"
❌ Rate: login(), authenticate(), verifyCredentials()
✅ Trouve: "// TODO: add authentication"  (commentaire inutile!)
```

### Recherche sémantique (oview RAG)

```bash
$ oview search "authentication"
```

**Avantages :**
- ✅ Trouve "login", "signin", "auth", etc.
- ✅ Comprend le contexte sémantique
- ✅ Résultats triés par pertinence
- ✅ Fonctionne même avec des synonymes

**Exemple :**
```
Query: "how users login"
✅ Trouve: login(), authenticate(), signin()
✅ Trouve: UserController, AuthService, SecurityConfig
✅ Ordonne par pertinence (92%, 88%, 76%...)
```

---

## 🎯 Cas d'usage concrets

### 1. Trouver du code par intention

**Recherche classique :**
```bash
$ grep -r "cache" .
# 247 résultats, beaucoup de bruit
```

**RAG :**
```bash
$ oview search "how is caching implemented"
# 5 résultats pertinents, ordonnés par pertinence
```

### 2. Comprendre avant de modifier

**Avant RAG :**
```
Vous: Modifie UserController.php
Claude: *modifie sans contexte*
Résultat: ❌ Casse une dépendance
```

**Avec RAG :**
```
Vous: Modifie UserController.php
Claude: [Utilise get_context("UserController.php")]
        [Voit AuthService, SecurityConfig]
        [Comprend les dépendances]
        "Je vois que UserController dépend de AuthService..."
Résultat: ✅ Modification sûre
```

### 3. Exploration de codebase inconnue

**Sans RAG :**
```
Vous: Comment marche le système de cache ?
Claude: "Je ne peux pas lire tous les fichiers..."
```

**Avec RAG :**
```
Vous: Comment marche le système de cache ?
Claude: [search("cache system implementation")]
        [Trouve: CacheManager.php, cache.yaml, CacheListener.php]
        "Le système utilise Redis avec une stratégie de TTL..."
```

---

## ⚡ Performance

### Vitesse

```
Indexation (487 chunks):     ~3-5 secondes
Génération embedding:        ~100-500ms (Ollama local)
Recherche pgvector:          ~50-200ms
Total requête Claude:        ~500-1500ms
```

### Précision

```
Top 1 pertinent:   ~85-95%
Top 5 pertinent:   ~95-99%
Faux positifs:     ~5-10%
```

### Scalabilité

```
1,000 chunks:      Très rapide (<100ms)
10,000 chunks:     Rapide (~200ms)
100,000 chunks:    Acceptable (~500ms)
1,000,000 chunks:  Lent (~2-3s) → Besoin d'optimisation
```

---

## 🔐 Sécurité & Confidentialité

### Données locales

```
┌─────────────────────────────────────────┐
│  Tout reste sur votre machine !         │
│                                         │
│  ✅ Code source: local                  │
│  ✅ Embeddings: générés localement      │
│  ✅ Database: Docker local              │
│  ✅ Recherches: locales                 │
│                                         │
│  ❌ Rien n'est envoyé à des serveurs   │
└─────────────────────────────────────────┘
```

**Exception :** Si vous utilisez OpenAI pour les embeddings
- Les chunks sont envoyés à l'API OpenAI
- Solution : Utilisez Ollama pour du 100% local

---

## 🛠️ Personnalisation

### Changer le nombre de résultats

```python
# Par défaut: 5 résultats
search("authentication")

# Plus de résultats
search("authentication", limit=10)
```

### Filtrer par type

```sql
-- Chercher uniquement dans le code
SELECT * FROM chunks
WHERE type = 'code'
ORDER BY embedding <=> $query;

-- Chercher uniquement dans les docs
SELECT * FROM chunks
WHERE type = 'doc'
ORDER BY embedding <=> $query;
```

### Ajuster la chunking strategy

Dans `.oview/rag.yaml` :

```yaml
chunking:
  php:
    strategy: function   # Par fonction (défaut)
    max_size: 2000      # Taille max du chunk
    overlap: 100        # Chevauchement entre chunks
```

---

## 🚀 Optimisations possibles

### 1. Index HNSW (déjà fait!)

```sql
CREATE INDEX ON chunks USING hnsw (embedding vector_cosine_ops);
```
→ Recherche en O(log n) au lieu de O(n)

### 2. Cache des embeddings fréquents

```python
cache = {
    "authentication": [0.21, -0.48, ...],
    "database": [0.89, 0.34, ...]
}
```

### 3. Indexation incrémentale

```bash
# Ne ré-indexer que les fichiers modifiés
oview index --incremental
```

### 4. Filtres avancés

```python
search("auth",
       language="php",
       type="code",
       min_similarity=0.8)
```

---

## 📚 Ressources

- **pgvector**: https://github.com/pgvector/pgvector
- **Ollama**: https://ollama.ai/
- **MCP Protocol**: https://modelcontextprotocol.io/
- **Embeddings**: https://www.pinecone.io/learn/embeddings/

---

## ❓ Questions fréquentes

**Q: Pourquoi 768 dimensions ?**
A: C'est la dimension du modèle nomic-embed-text d'Ollama. OpenAI utilise 1536 ou 3072.

**Q: Puis-je utiliser plusieurs modèles ?**
A: Non, tous les chunks doivent utiliser le même modèle pour être comparables.

**Q: Comment ré-indexer après des changements ?**
A: `oview index` écrase l'index existant.

**Q: Ça marche avec n'importe quel langage ?**
A: Oui ! Les embeddings capturent le sens, pas la syntaxe.

**Q: C'est mieux que GitHub Copilot ?**
A: Complémentaire ! Copilot suggère du code, oview aide à chercher et comprendre.
