# Guide des Embeddings pour oview

## Vue d'ensemble

Les **embeddings** sont des vecteurs qui capturent le **sens sémantique** du code. Ils permettent de faire des recherches intelligentes comme "comment l'authentification fonctionne ?" au lieu de simplement chercher le mot "auth".

## Options disponibles

### 1. Stub (par défaut - MVP) ⚠️

**Utilisation :**
```bash
oview index
# ou explicitement
oview index --embeddings=stub
```

**Caractéristiques :**
- ❌ **Aucune compréhension sémantique**
- ✅ Gratuit et rapide
- ✅ Pas besoin d'API ou de configuration
- ✅ Déterministe (même texte = même vecteur)

**Quand l'utiliser :** Tests et développement de l'infrastructure uniquement.

---

### 2. OpenAI (Recommandé) ✅

**Modèles disponibles :**
- `text-embedding-3-small` (défaut) - **$0.02 / 1M tokens** ⭐ Recommandé
- `text-embedding-3-large` - $0.13 / 1M tokens (meilleur qualité)
- `text-embedding-ada-002` - $0.10 / 1M tokens (ancien)

**Installation :**
1. Créez un compte sur https://platform.openai.com
2. Obtenez votre clé API
3. Configurez la variable d'environnement :

```bash
export OPENAI_API_KEY="sk-..."
```

Ou ajoutez dans `~/.zshrc` :
```bash
echo 'export OPENAI_API_KEY="sk-..."' >> ~/.zshrc
source ~/.zshrc
```

**Utilisation :**

```bash
# Via variable d'environnement (recommandé)
oview index --embeddings=openai

# Via flag
oview index --embeddings=openai --openai-key="sk-..."

# Avec un modèle spécifique
OPENAI_MODEL="text-embedding-3-large" oview index --embeddings=openai
```

**Coût estimé pour votre projet (7840 chunks) :**
- Avec text-embedding-3-small : **~$0.12** (5.7 MB ≈ 1.4M tokens)
- Indexation complète : quelques minutes
- Réindexation incrémentale : quasi gratuite

**Avantages :**
- ✅ Qualité exceptionnelle
- ✅ Très rapide (API)
- ✅ Pas d'infrastructure à gérer
- ✅ Support multilingue (PHP, JS, commentaires en français)

---

### 3. Ollama (Local - Gratuit) 🏠

**Installation d'Ollama :**

```bash
# Linux
curl -fsSL https://ollama.com/install.sh | sh

# macOS
brew install ollama

# Démarrer le service
ollama serve
```

**Télécharger un modèle :**

```bash
# Modèle recommandé (768 dimensions, 274 MB)
ollama pull nomic-embed-text

# Alternatives
ollama pull mxbai-embed-large  # 1024 dimensions, 669 MB
ollama pull all-minilm         # 384 dimensions, 45 MB (plus rapide)
```

**Utilisation :**

```bash
# Avec le modèle par défaut (nomic-embed-text)
oview index --embeddings=ollama

# Avec un modèle spécifique
oview index --embeddings=ollama --ollama-model=mxbai-embed-large

# Avec une URL custom
oview index --embeddings=ollama --ollama-url=http://localhost:11434
```

**Note importante sur les dimensions :**
Si vous utilisez Ollama, vous devrez **adapter le schéma de la base** :

```bash
# 1. Vérifier la dimension du modèle
ollama show nomic-embed-text | grep -i dimension
# Output: 768 dimensions

# 2. Adapter la table chunks
docker exec oview-postgres psql -U oview -d oview_chapitreneuf -c \
  "ALTER TABLE chunks ALTER COLUMN embedding TYPE vector(768);"

# 3. Réindexer
oview index --embeddings=ollama
```

**Avantages :**
- ✅ Gratuit et privé
- ✅ Pas besoin d'Internet
- ✅ Données ne quittent pas votre machine
- ✅ Pas de limite de tokens

**Inconvénients :**
- ⚠️ Plus lent que l'API OpenAI
- ⚠️ Qualité légèrement inférieure
- ⚠️ Nécessite de la RAM (2-4 GB)

---

## Comparaison rapide

| Critère | Stub | OpenAI | Ollama |
|---------|------|--------|--------|
| **Qualité** | ❌ Aucune | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ |
| **Coût** | Gratuit | ~$0.10 | Gratuit |
| **Vitesse** | ⚡⚡⚡ | ⚡⚡⚡ | ⚡⚡ |
| **Setup** | Aucun | Clé API | Installation |
| **Privé** | ✅ | ❌ | ✅ |
| **Internet** | Non | Oui | Non |
| **Recommandé pour** | Tests infra | Production | Dev local |

## Workflow recommandé

### Pour le développement (votre cas actuel)

**Option 1 : OpenAI (rapide à tester)**
```bash
# 1. Configurez la clé
export OPENAI_API_KEY="sk-..."

# 2. Réindexez avec de vrais embeddings
cd ~/Documents/chapitreneuf
oview index --embeddings=openai

# Coût : ~$0.12 pour vos 797 fichiers
```

**Option 2 : Ollama (gratuit mais plus d'installation)**
```bash
# 1. Installez Ollama
curl -fsSL https://ollama.com/install.sh | sh

# 2. Démarrez le service
ollama serve &

# 3. Téléchargez le modèle
ollama pull nomic-embed-text

# 4. Adaptez la base (768 dimensions au lieu de 1536)
docker exec oview-postgres psql -U oview -d oview_chapitreneuf -c \
  "ALTER TABLE chunks ALTER COLUMN embedding TYPE vector(768);"

# 5. Réindexez
cd ~/Documents/chapitreneuf
oview index --embeddings=ollama
```

### Pour la production

Utilisez **OpenAI** pour la qualité et la simplicité, avec un mécanisme de cache pour minimiser les coûts :
- Indexation complète : une fois
- Réindexations : uniquement les fichiers modifiés (TODO: à implémenter)

---

## Vérification

Après réindexation avec de vrais embeddings, vérifiez :

```bash
# Comptez les chunks
docker exec oview-postgres psql -U oview -d oview_chapitreneuf -c \
  "SELECT COUNT(*) FROM chunks;"

# Vérifiez qu'il y a bien des embeddings non-nuls
docker exec oview-postgres psql -U oview -d oview_chapitreneuf -c \
  "SELECT path, LENGTH(embedding::text) as embedding_size FROM chunks LIMIT 5;"
```

Les embeddings OpenAI/Ollama auront une taille ~20-30KB (texte du vecteur), alors que les stubs sont plus courts.

---

## Dépannage

### "OpenAI API error: 401"
→ Clé API invalide. Vérifiez votre `OPENAI_API_KEY`

### "Ollama API request failed: connection refused"
→ Ollama n'est pas démarré. Lancez `ollama serve`

### "Model not found"
→ Le modèle n'est pas téléchargé. Lancez `ollama pull nomic-embed-text`

### Embeddings trop longs (dépassement de token limit)
→ Les chunks sont automatiquement tronqués à 30k caractères

---

## Performance

### Temps d'indexation estimé pour 797 fichiers (7840 chunks)

- **Stub** : ~1 minute (votre temps actuel)
- **OpenAI** : ~5-10 minutes (limité par l'API rate limit)
- **Ollama** : ~15-30 minutes (dépend de votre CPU/GPU)

### Optimisations futures possibles
- Indexation parallèle (batch requests)
- Cache des embeddings par hash de contenu
- Indexation incrémentale (uniquement fichiers modifiés)

---

## Prochaines étapes

Une fois les vrais embeddings en place, vous pourrez :
1. Implémenter un système de requêtes RAG
2. Faire des recherches sémantiques : "comment fonctionne l'authentification ?"
3. Utiliser les agents Claude avec contexte RAG pertinent
4. Construire des workflows n8n qui utilisent le contexte du code

**Voulez-vous que je vous aide à implémenter un de ces points ?**
