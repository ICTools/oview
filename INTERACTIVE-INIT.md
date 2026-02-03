# Guide d'Initialisation Interactive

## Vue d'ensemble

`oview init` est maintenant **interactif** ! Plus besoin d'éditer manuellement `.oview/project.yaml`.

## Workflow

```bash
cd ~/Documents/mon-projet
oview init
```

### Étape 1 : Détection automatique

```
🔍 Initializing oview for this project...

📁 Creating .oview directory structure...
   ✓ Directory structure created
🔎 Detecting project stack...
   ✓ Stack detected:
     - Symfony
     - Docker
     - Makefile
     - Frontend: [Symfony UX]
     - Languages: [PHP JavaScript]
```

### Étape 2 : Configuration des embeddings

```
🤖 Configuration interactive

📊 Configuration des embeddings (vecteurs sémantiques)

Les embeddings permettent la recherche sémantique dans votre code.

Providers disponibles:
  1. stub         - Placeholder (hash, pas de sémantique) - Gratuit
  2. openai       - OpenAI API (haute qualité) - ~$0.02/1M tokens
  3. ollama       - Local (privé, gratuit) - Nécessite installation

Choisir provider [1-3] (défaut: 1):
```

**Choix recommandés :**

#### Pour développement local (privé, gratuit)
```
Choisir provider [1-3]: 3

Modèles Ollama populaires:
  1. nomic-embed-text   - 768 dim, 274 MB (recommandé)
  2. mxbai-embed-large  - 1024 dim, 669 MB
  3. bge-code           - 768 dim, optimisé code
  4. all-minilm         - 384 dim, 45 MB (rapide)

Choisir modèle [1-4]: 1
Base URL Ollama (défaut: http://localhost:11434): [Enter]

💡 Avant d'indexer, lancez: ollama serve && ollama pull nomic-embed-text
```

#### Pour production (qualité maximale)
```
Choisir provider [1-3]: 2

Modèles OpenAI disponibles:
  1. text-embedding-3-small  - $0.02/1M tokens, 1536 dim (recommandé)
  2. text-embedding-3-large  - $0.13/1M tokens, 3072 dim (meilleure qualité)
  3. text-embedding-ada-002  - $0.10/1M tokens, 1536 dim (ancien)

Choisir modèle [1-3]: 1

💡 N'oubliez pas de configurer OPENAI_API_KEY dans votre environnement
```

#### Pour tests (sans setup)
```
Choisir provider [1-3]: 1

ℹ️  Stub: Pas de sémantique, uniquement pour tester l'infrastructure
```

### Étape 3 : Configuration du LLM

```
🤖 Configuration du LLM (agent AI)

Le LLM sera utilisé par les agents pour analyser et modifier le code.

Providers disponibles:
  1. claude-code   - Claude Code CLI (Sonnet 4.5) - Intégré
  2. claude-api    - Claude API (Anthropic) - Nécessite clé API
  3. openai        - OpenAI API (GPT-4, etc.) - Nécessite clé API
  4. ollama        - Local (Llama 3, etc.) - Gratuit

Choisir provider [1-4] (défaut: 1):
```

**Choix recommandés :**

#### Claude Code (défaut, recommandé)
```
Choisir provider [1-4]: 1

✅ Claude Code: Utilise le CLI actuel (recommandé)
```

#### Claude API (si vous préférez l'API)
```
Choisir provider [1-4]: 2

Modèles Claude API:
  1. claude-sonnet-4.5    - Dernier, équilibré (recommandé)
  2. claude-opus-4.5      - Maximum qualité
  3. claude-haiku-4       - Rapide et économique

Choisir modèle [1-3]: 1

💡 Configurez ANTHROPIC_API_KEY dans votre environnement
```

#### Ollama (local, gratuit)
```
Choisir provider [1-4]: 4

Modèles Ollama populaires:
  1. llama3.1:70b      - Haute qualité
  2. llama3.1:8b       - Rapide
  3. codellama:34b     - Optimisé code
  4. deepseek-coder    - Spécialisé code

Choisir modèle [1-4]: 2
Base URL Ollama (défaut: http://localhost:11434): [Enter]

💡 Avant d'utiliser, lancez: ollama serve && ollama pull llama3.1:8b
```

### Étape 4 : Finalisation

```
📝 Creating project configuration...
   ✓ Project config saved (slug: mon-projet)
📋 Creating RAG configuration...
   ✓ RAG config saved
📊 Creating index manifests...
   ✓ Index manifests created
🤖 Generating Claude agent instruction files...
   ✓ Agent files generated

✅ Initialization complete!
```

## Résultat dans `.oview/project.yaml`

```yaml
project_id: abc123
project_slug: mon-projet
embeddings:
  provider: ollama
  model: nomic-embed-text
  dim: 768
  base_url: http://localhost:11434
llm:
  provider: claude-code
  model: claude-sonnet-4.5
```

## Mode non-interactif

Pour les scripts et CI/CD :

```bash
oview init --non-interactive
```

Utilise les valeurs par défaut :
- Embeddings : stub
- LLM : claude-code (Sonnet 4.5)

## Reconfiguration

Si vous voulez changer la config après coup :

### Option 1 : Réinitialiser (recommandé)
```bash
oview init --force
# Répond aux questions interactives
```

### Option 2 : Édition manuelle
```bash
vim .oview/project.yaml
# Modifiez les sections embeddings et llm
```

Puis :
```bash
# Si vous avez changé le modèle d'embeddings
oview index  # Réindexe avec le nouveau modèle

# Si vous avez changé le LLM
# Rien à faire, il sera utilisé au prochain appel d'agent
```

## Exemples de combinaisons

### Dev local full open-source
```
Embeddings: ollama / nomic-embed-text
LLM:        ollama / llama3.1:8b
```

**Setup :**
```bash
ollama serve &
ollama pull nomic-embed-text
ollama pull llama3.1:8b
```

**Avantages :**
- ✅ 100% gratuit
- ✅ 100% privé
- ✅ Pas besoin d'Internet

**Inconvénients :**
- ⚠️ Nécessite RAM/GPU
- ⚠️ Plus lent

### Production qualité maximale
```
Embeddings: openai / text-embedding-3-small
LLM:        claude-code / claude-sonnet-4.5
```

**Setup :**
```bash
export OPENAI_API_KEY="sk-..."
# Claude Code déjà configuré
```

**Avantages :**
- ✅ Qualité maximale
- ✅ Rapide
- ✅ Claude Code intégré

**Inconvénients :**
- 💰 ~$0.02 par 1M tokens embeddings
- 💰 Coût Claude selon usage

### Compromis (recommandé)
```
Embeddings: ollama / nomic-embed-text  (local, gratuit)
LLM:        claude-code / claude-sonnet-4.5  (intégré)
```

**Setup :**
```bash
ollama serve &
ollama pull nomic-embed-text
# Claude Code déjà configuré
```

**Avantages :**
- ✅ Embeddings gratuits et privés
- ✅ LLM de haute qualité
- ✅ Bon équilibre

### Tests et développement infra
```
Embeddings: stub / stub-hash-based
LLM:        claude-code / claude-sonnet-4.5
```

**Aucun setup nécessaire !**

Parfait pour :
- Tester l'infrastructure
- Développer des features
- CI/CD

## Navigation interactive

**Entrée vide = valeur par défaut**
```
Choisir provider [1-3] (défaut: 1): [Enter]
→ Utilise option 1
```

**Numéro OU nom**
```
Choisir modèle [1-4]: 3
→ Utilise option 3

Choisir modèle [1-4]: nomic-embed-text
→ Trouve et utilise nomic-embed-text
```

**Case insensitive**
```
Choisir provider: OPENAI
→ Fonctionne
```

## Vérification de la config

```bash
# Voir la config complète
cat .oview/project.yaml

# Voir uniquement embeddings
cat .oview/project.yaml | grep -A 5 "embeddings:"

# Voir uniquement LLM
cat .oview/project.yaml | grep -A 4 "llm:"
```

## Intégration avec le reste du workflow

### Après init avec Ollama embeddings

```bash
# 1. Init (fait)
oview init
# → Choisit ollama / nomic-embed-text

# 2. Setup Ollama
ollama serve &
ollama pull nomic-embed-text

# 3. Adapter la DB pour 768 dimensions
oview up
docker exec oview-postgres psql -U oview -d oview_mon-projet -c \
  "ALTER TABLE chunks ALTER COLUMN embedding TYPE vector(768);"

# 4. Indexer
oview index
```

### Après init avec OpenAI embeddings

```bash
# 1. Init (fait)
oview init
# → Choisit openai / text-embedding-3-small

# 2. Configurer la clé
export OPENAI_API_KEY="sk-..."

# 3. Setup DB
oview up

# 4. Indexer
oview index
```

### Après init avec stub

```bash
# 1. Init (fait)
oview init
# → Choisit stub

# 2. Setup DB
oview up

# 3. Indexer (stub, rapide)
oview index

# 4. Plus tard, passer à de vrais embeddings
vim .oview/project.yaml
# Changer provider + model + dim

# 5. Réindexer
oview index
```

## FAQ

### Puis-je sauter l'interactif ?

Oui :
```bash
oview init --non-interactive
```

### Puis-je relancer init après ?

Oui :
```bash
oview init --force
```

Répond aux questions, écrase la config.

### Les choix sont-ils validés ?

Oui, seuls les choix valides sont acceptés. En cas d'erreur, la valeur par défaut est utilisée.

### Puis-je éditer manuellement après ?

Oui, `.oview/project.yaml` reste éditable.

### Que se passe-t-il si j'appuie juste sur Entrée ?

La valeur par défaut est utilisée (indiquée entre parenthèses).

### Les API keys sont-elles stockées ?

**Non** (par défaut). Le champ `api_key` existe mais est vide.

**Recommandation :** Utilisez les variables d'environnement :
- `OPENAI_API_KEY`
- `ANTHROPIC_API_KEY`

Si vous voulez vraiment stocker dans le fichier :
```yaml
embeddings:
  api_key: sk-...  # ⚠️ Ne commitez JAMAIS ce fichier avec une clé !
```

### Puis-je utiliser des modèles custom ?

Oui ! Si vous tapez un nom qui n'est pas dans la liste, il sera utilisé tel quel.

Exemple :
```
Choisir modèle: mon-modele-custom
→ Utilise "mon-modele-custom"
```

---

**L'init interactif rend oview accessible aux débutants tout en restant flexible pour les experts !** 🚀
