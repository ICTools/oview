# Embeddings - Workflow Simple

## Philosophie

**Une seule source de vérité** : `.oview/project.yaml`

Tu configures une fois, tu indexes autant de fois que tu veux.

## 1. Configuration

Édite `.oview/project.yaml` :

```yaml
embeddings:
  provider: stub      # stub, openai, ou ollama
  model: stub-hash-based
  dim: 1536
```

### Pour OpenAI

```yaml
embeddings:
  provider: openai
  model: text-embedding-3-small  # ou text-embedding-3-large
  dim: 1536
  # api_key: sk-...  # optionnel, préfère OPENAI_API_KEY env var
```

### Pour Ollama (local)

```yaml
embeddings:
  provider: ollama
  model: nomic-embed-text  # ou bge-code, mxbai-embed-large
  dim: 768                 # attention: 768 pour nomic, 1024 pour mxbai
  base_url: http://localhost:11434  # optionnel
```

### Pour BGE Code (exemple)

```yaml
embeddings:
  provider: ollama
  model: bge-code
  dim: 768
```

## 2. Indexation

```bash
cd ~/Documents/chapitreneuf
oview index
```

**C'est tout.** Il lit la config, génère les embeddings, stocke dans la DB.

## 3. Changer de modèle

Tu veux passer de `stub` à `bge-code` ?

### Étape 1 : Édite le fichier

```yaml
embeddings:
  provider: ollama
  model: bge-code
  dim: 768
```

### Étape 2 : Adapte la DB (si dimension change)

```bash
# Si tu passes de 1536 → 768
docker exec oview-postgres psql -U oview -d oview_chapitreneuf -c \
  "ALTER TABLE chunks ALTER COLUMN embedding TYPE vector(768);"
```

### Étape 3 : Réindex

```bash
oview index
```

**C'est normal et sain** :
- ✅ Vide les vieux embeddings
- ✅ Régénère avec le nouveau modèle
- ✅ Même texte, nouvelle "carte sémantique"

## 4. Vérification

```bash
# Check quel modèle est utilisé
docker exec oview-postgres psql -U oview -d oview_chapitreneuf -c \
  "SELECT DISTINCT embedding_model FROM chunks;"

# Output attendu:
#  embedding_model
# -----------------
#  bge-code
```

## 5. Exemples de configs

### Dev local (gratuit, privé)

```yaml
embeddings:
  provider: ollama
  model: nomic-embed-text
  dim: 768
```

Avant la première indexation :
```bash
ollama serve &
ollama pull nomic-embed-text
```

### Production (qualité max)

```yaml
embeddings:
  provider: openai
  model: text-embedding-3-small
  dim: 1536
```

Puis :
```bash
export OPENAI_API_KEY="sk-..."
oview index
```

### Tests (pas de coût, pas de setup)

```yaml
embeddings:
  provider: stub
  model: stub-hash-based
  dim: 1536
```

## Architecture DB

```sql
CREATE TABLE chunks (
    ...
    embedding vector(1536),
    embedding_model VARCHAR(100),  -- Stocke "bge-code", "text-embedding-3-small", etc.
    ...
);
```

**Pourquoi ?**
- Tu peux mixer plusieurs modèles (migration incrémentale)
- Tu sais toujours quelle version de "carte" tu utilises
- Tu peux filtrer par modèle dans tes requêtes

## Migration entre modèles

### Scénario : Tu as indexé avec stub, tu veux passer à OpenAI

**Avant :**
```sql
SELECT embedding_model, COUNT(*) FROM chunks GROUP BY embedding_model;
-- stub-hash-based | 7840
```

**Tu changes la config :**
```yaml
embeddings:
  provider: openai
  model: text-embedding-3-small
  dim: 1536  # même dimension, pas besoin d'ALTER TABLE
```

**Tu réindexes :**
```bash
oview index
```

**Après :**
```sql
SELECT embedding_model, COUNT(*) FROM chunks GROUP BY embedding_model;
-- text-embedding-3-small | 7840
```

**Propre. Simple. Tracé.**

## Commandes utiles

### Voir la config actuelle
```bash
cat .oview/project.yaml | grep -A 5 "embeddings:"
```

### Compter les chunks par modèle
```bash
docker exec oview-postgres psql -U oview -d oview_chapitreneuf -c \
  "SELECT embedding_model, COUNT(*) FROM chunks GROUP BY embedding_model;"
```

### Vider et réindexer (force clean)
```bash
# Méthode 1 : via SQL
docker exec oview-postgres psql -U oview -d oview_chapitreneuf -c \
  "DELETE FROM chunks WHERE project_id = '$(cat .oview/project.yaml | grep project_id | cut -d: -f2 | tr -d ' ')';"

# Méthode 2 : via oview (fait automatiquement à chaque index)
oview index
```

## Workflow recommandé

1. **Init** : `oview init` → crée config avec stub
2. **Test infra** : `oview index` → vérifie que tout fonctionne
3. **Config embeddings** : édite `.oview/project.yaml`
4. **Setup** (si Ollama) : `ollama pull ton-modele`
5. **Index réel** : `oview index`
6. **Query** (futur) : `oview query "comment marche l'auth ?"`

## Pas de flags, pas de confusion

❌ **Avant (complexe)** :
```bash
oview index --embeddings=openai --openai-key=... --model=...
```

✅ **Maintenant (simple)** :
```bash
# 1. Édite le fichier
vim .oview/project.yaml

# 2. Index
oview index
```

**Un fichier. Une commande. Zéro ambiguïté.**
