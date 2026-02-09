# 📋 TODO - oview Improvements

**Format:** Each task is written as a prompt you can copy/paste to iterate

---

## 🔥 P0 - CRITICAL (Do First)

### 1. Support multi-langages pour Tree-sitter chunking

**Prompt:**
> Ajoute le support complet des langages suivants dans Tree-sitter :
> - Java (classes, méthodes, interfaces)
> - Rust (modules, fonctions, traits, impl blocks)
> - C/C++ (fonctions, classes, structs)
> - Ruby (classes, méthodes, modules)
> - C# (classes, méthodes, namespaces)
>
> Pour chaque langage :
> 1. Ajoute la dépendance go-tree-sitter dans `go.mod`
> 2. Importe le parser dans `internal/treesitter/parser.go`
> 3. Ajoute le case dans `getLanguage()`
> 4. Ajoute dans `SupportedLanguages()`
> 5. Crée les query patterns dans `internal/treesitter/extractor.go`
> 6. Teste avec des fichiers réels

**Context:**
- Tree-sitter parsers disponibles : https://github.com/smacker/go-tree-sitter
- Parsers actuellement supportés : Python, JS, TS, Go, PHP (5 langages)
- Pattern existant dans `extractor.go` pour extraire fonctions/classes

**Files to modify:**
- `internal/treesitter/parser.go` (add language cases)
- `internal/treesitter/extractor.go` (add extraction patterns)
- `go.mod` (add tree-sitter dependencies)
- `internal/indexer/chunker.go` (update detectLanguage())

**Acceptance Criteria:**
- [ ] 10+ langages supportés (vs 5 actuellement)
- [ ] Extraction de fonctions/classes/méthodes pour chaque langage
- [ ] Tests avec fichiers réels de chaque langage
- [ ] Documentation des patterns d'extraction dans TREE_SITTER_CHUNKING.md
- [ ] Commande `oview index` fonctionne sur projets multi-langages

**Estimated effort:** 1 jour

**Priority justification:** Impact majeur sur l'utilité d'oview pour projets polyglots

---

### 2. Corriger le bug SQL dans MCP search

**Prompt:**
> Le serveur MCP échoue avec l'erreur "pq: got 3 parameters but the statement requires 2".
>
> Debug et corrige :
> 1. Identifie la requête SQL problématique dans `internal/mcp/handler.go`
> 2. Compte les placeholders ($1, $2, ...) dans la query
> 3. Compte les paramètres passés dans `db.Query()` ou `db.Exec()`
> 4. Corrige le mismatch
> 5. Teste avec `oview mcp` et une recherche via Claude Code

**Context:**
- Erreur actuelle : `MCP error -32000: search failed: query failed: pq: got 3 parameters but the statement requires 2`
- Le MCP search fonctionne parfois puis échoue
- Probablement lié aux filtres optionnels (type, language, path_pattern, etc.)

**Files to check:**
- `internal/mcp/handler.go:searchSimilarChunks()`
- `internal/query/filters.go`

**Acceptance Criteria:**
- [ ] MCP search fonctionne sans erreur SQL
- [ ] Tous les filtres (type, language, component, path) fonctionnent
- [ ] Test avec différentes combinaisons de filtres
- [ ] Logs MCP montrent requêtes réussies

**Estimated effort:** 2 heures

---

### 3. Ajouter tests unitaires critiques

**Prompt:**
> Actuellement il n'y a qu'1 seul test sur 28 fichiers Go (3.6% coverage).
>
> Ajoute des tests pour les packages critiques :
>
> **internal/embeddings/**
> - Test `CountTokens()` pour Ollama, OpenAI, Stub
> - Test `Embed()` avec truncation
> - Test fallback quand tokenizer fail
> - Mock HTTP pour Ollama API
>
> **internal/indexer/**
> - Test chunking de différents types de fichiers
> - Test indexation complète avec mock DB
> - Test détection de langages
>
> **internal/mcp/**
> - Test handlers (search, get_context, project_info)
> - Test parsing des filtres
> - Test stratégies de recherche
>
> **internal/treesitter/**
> - Test extraction pour chaque langage
> - Test chunking avec limites de tokens
> - Test fallback quand parsing échoue

**Context:**
- Seul test existant : `internal/claude/rag_policy_test.go`
- Framework : Go testing standard + testify (à ajouter)
- Besoin de mocks pour : DB, HTTP, embeddings API

**Files to create:**
- `internal/embeddings/ollama_test.go`
- `internal/embeddings/openai_test.go`
- `internal/indexer/chunker_test.go`
- `internal/indexer/indexer_test.go`
- `internal/mcp/handler_test.go`
- `internal/treesitter/chunker_test.go`
- `internal/treesitter/extractor_test.go`

**Acceptance Criteria:**
- [ ] Coverage > 60% sur packages critiques
- [ ] Tests passent : `go test ./... -v`
- [ ] CI/CD configuré (GitHub Actions)
- [ ] Tests rapides (< 5s total)

**Estimated effort:** 3 jours

---

## 🔴 P1 - HIGH (Next Sprint)

### 4. Ajouter gestion de contexte Go

**Prompt:**
> L'application n'utilise pas le système de contexte Go, ce qui empêche :
> - L'annulation des opérations longues
> - Les timeouts sur appels API
> - La propagation de deadlines
>
> Ajoute context.Context à toutes les opérations I/O :
> 1. Modifie l'interface `Generator` : `Embed(ctx context.Context, text string)`
> 2. Propage context dans indexer, chunker, MCP handlers
> 3. Ajoute timeouts sur :
>    - Appels Ollama API (30s)
>    - Appels OpenAI API (30s)
>    - Requêtes DB (10s)
>    - Indexation complète (peut continuer)
> 4. Gère context.Canceled proprement

**Files to modify:**
- `internal/embeddings/interface.go`
- `internal/embeddings/ollama.go`
- `internal/embeddings/openai.go`
- `internal/indexer/indexer.go`
- `internal/mcp/handler.go`
- `cmd/index.go` (context.WithCancel sur SIGINT)

**Acceptance Criteria:**
- [ ] Toutes les opérations I/O acceptent context
- [ ] CTRL+C annule l'indexation proprement
- [ ] Timeouts configurables via config
- [ ] Tests avec context.WithTimeout()

**Estimated effort:** 1 jour

---

### 5. Implémenter cache d'embeddings

**Prompt:**
> Les requêtes de recherche identiques régénèrent les embeddings à chaque fois.
>
> Implémente un cache LRU :
> 1. Crée `internal/cache/lru.go` avec interface simple :
>    ```go
>    type Cache interface {
>        Get(key string) ([]float32, bool)
>        Set(key string, value []float32)
>    }
>    ```
> 2. Implémente avec `github.com/hashicorp/golang-lru`
> 3. Hash la query pour générer la clé
> 4. Cache dans MCP handler avant appel `Embed()`
> 5. Ajoute métriques : hit rate, size
> 6. Config : max_size (default 1000), TTL (default 1h)

**Files to create:**
- `internal/cache/lru.go`
- `internal/cache/lru_test.go`

**Files to modify:**
- `internal/mcp/handler.go` (wrap generator avec cache)
- `.oview/project.yaml` (config cache)

**Acceptance Criteria:**
- [ ] Requêtes identiques utilisent le cache
- [ ] Hit rate > 80% en usage normal
- [ ] Latence divisée par 10 sur cache hit
- [ ] Métriques exposées dans logs

**Estimated effort:** 2 jours

---

### 6. Optimiser tokenizer Ollama (éliminer binary search)

**Prompt:**
> La troncature actuelle fait log(N) appels API pour trouver la bonne longueur.
>
> Optimise en :
> 1. Faisant 1 seul appel `CountTokens()` sur le texte complet
> 2. Si > max, estime le ratio chars/token réel : `ratio = len(text) / tokenCount`
> 3. Tronque à `text[:(maxTokens * ratio * 0.95)]` (5% marge)
> 4. Vérifie avec 1 seul appel `CountTokens()`
> 5. Ajuste si nécessaire (rare)
>
> Bénéfice : 1-2 appels au lieu de 10-15

**Files to modify:**
- `internal/embeddings/ollama.go:truncateToTokenLimit()`

**Acceptance Criteria:**
- [ ] Maximum 2 appels CountTokens() par truncation
- [ ] Tests vérifient pas d'appels excessifs
- [ ] Latence divisée par 5-10x

**Estimated effort:** 3 heures

---

### 7. Paralléliser l'indexation

**Prompt:**
> L'indexation traite 53 fichiers en 26s séquentiellement.
>
> Parallélise avec worker pool :
> 1. Crée `internal/indexer/workers.go` :
>    ```go
>    type WorkerPool struct {
>        workers int
>        jobs chan Job
>        results chan Result
>    }
>    ```
> 2. Nombre de workers = `runtime.NumCPU()`
> 3. Channel de jobs (files à indexer)
> 4. Channel de results (chunks + erreurs)
> 5. Barre de progression avec atomic counter
> 6. Gestion d'erreurs : continue sur erreur, collecte toutes les erreurs
> 7. Configuration : `--workers N` flag

**Files to create:**
- `internal/indexer/workers.go`
- `internal/indexer/workers_test.go`

**Files to modify:**
- `internal/indexer/indexer.go:Index()`
- `cmd/index.go` (add --workers flag)

**Acceptance Criteria:**
- [ ] Indexation 4-8x plus rapide (53 files en 3-6s)
- [ ] Pas de race conditions
- [ ] Erreurs bien gérées et reportées
- [ ] Tests avec différents nombre de workers

**Estimated effort:** 2 jours

---

## 🟠 P2 - MEDIUM (Future)

### 8. Indexation incrémentale

**Prompt:**
> Actuellement `oview index` réindexe tous les fichiers.
>
> Implémente l'indexation incrémentale :
> 1. Calcule SHA256 de chaque fichier
> 2. Compare avec `.oview/index/manifest.json`
> 3. Identifie : nouveaux, modifiés, supprimés
> 4. Indexe seulement les changés
> 5. Supprime chunks des fichiers deleted
> 6. Flag `--full` pour forcer full reindex
> 7. Ajoute métrique : files_skipped

**Files to modify:**
- `internal/indexer/indexer.go`
- `.oview/index/manifest.json` (add file hashes)
- `cmd/index.go` (--full flag)

**Acceptance Criteria:**
- [ ] 90% de temps économisé sur petit changement
- [ ] Manifest correctement mis à jour
- [ ] Chunks orphelins nettoyés
- [ ] Tests avec scénarios: add, modify, delete

**Estimated effort:** 3 jours

---

### 9. Métriques Prometheus

**Prompt:**
> Ajoute observabilité avec Prometheus :
> 1. Crée `internal/metrics/prometheus.go`
> 2. Métriques à exposer :
>    - `oview_index_duration_seconds` (histogram)
>    - `oview_index_files_total` (counter)
>    - `oview_index_chunks_total` (counter)
>    - `oview_mcp_requests_total` (counter by tool)
>    - `oview_mcp_request_duration_seconds` (histogram)
>    - `oview_cache_hits_total` / `cache_misses_total`
>    - `oview_embeddings_calls_total` (counter by provider)
> 3. Endpoint HTTP : `oview serve --metrics-port 9090`
> 4. Instrumente code avec `prometheus.NewHistogramVec()` etc.

**Files to create:**
- `internal/metrics/prometheus.go`
- `cmd/serve.go` (new command)

**Files to modify:**
- `internal/indexer/indexer.go` (record metrics)
- `internal/mcp/handler.go` (record metrics)

**Acceptance Criteria:**
- [ ] Endpoint /metrics exposé
- [ ] Grafana dashboard fourni
- [ ] Documentation de setup Prometheus

**Estimated effort:** 2 jours

---

### 10. Structured logging avec contexte

**Prompt:**
> Les logs actuels sont basiques. Améliore avec :
> 1. Remplace fmt.Printf par `zerolog` ou `slog`
> 2. Structured logging : JSON en production, human-readable en dev
> 3. Niveaux : DEBUG, INFO, WARN, ERROR
> 4. Contexte : request_id, user, operation
> 5. Config : `--log-level` flag, `OVIEW_LOG_LEVEL` env var
> 6. Logs MCP avec correlation IDs

**Files to create:**
- `internal/logger/logger.go` (déjà existe, améliorer)

**Files to modify:**
- Tous les fichiers avec `fmt.Printf`
- `internal/mcp/server.go` (add correlation ID)

**Acceptance Criteria:**
- [ ] Logs structurés en JSON
- [ ] Niveaux configurables
- [ ] Correlation IDs dans MCP
- [ ] Rotation des logs

**Estimated effort:** 1 jour

---

### 11. File watcher pour auto-reindex

**Prompt:**
> Ajoute un mode watch qui réindexe automatiquement :
> 1. Use `github.com/fsnotify/fsnotify`
> 2. Nouvelle commande : `oview watch`
> 3. Watch les dossiers de `.oview/rag.yaml:indexing.include_paths`
> 4. Debounce 2s (plusieurs changements = 1 reindex)
> 5. Ignore `.git/`, `node_modules/`, `.oview/`
> 6. Log chaque reindex déclenché

**Files to create:**
- `cmd/watch.go`
- `internal/watcher/watcher.go`

**Acceptance Criteria:**
- [ ] Détecte changements de fichiers
- [ ] Debouncing fonctionne
- [ ] Performance acceptable (pas de CPU spike)
- [ ] Peut être stoppé proprement (CTRL+C)

**Estimated effort:** 1 jour

---

## 🟢 P3 - LOW (Backlog)

### 12. Web UI pour gestion d'index

**Prompt:**
> Crée une UI web simple :
> - Backend : Go HTTP server avec API REST
> - Frontend : HTML/CSS/JS vanilla (pas de framework)
> - Features :
>   - Vue des chunks indexés (table paginée)
>   - Recherche sémantique interactive
>   - Stats d'indexation (graphiques)
>   - Logs en temps réel (SSE)
>   - Trigger reindex manuel
> - Route : `oview serve --port 8080`

**Estimated effort:** 1 semaine

---

### 13. Migrations DB avec golang-migrate

**Prompt:**
> Ajoute système de migrations :
> 1. Use `github.com/golang-migrate/migrate`
> 2. Migrations dans `migrations/`
> 3. Commande : `oview migrate up/down`
> 4. Version tracking en DB
> 5. Auto-migrate au démarrage (opt-in)

**Estimated effort:** 1 jour

---

### 14. Support Docker Compose pour dev

**Prompt:**
> Crée docker-compose.yml pour dev :
> - Service oview (rebuild on change)
> - Service postgres+pgvector
> - Service ollama (optionnel)
> - Volumes pour persistence
> - Health checks

**Estimated effort:** 4 heures

---

### 15. CLI auto-completion

**Prompt:**
> Génère shell completion :
> - `oview completion bash/zsh/fish`
> - Cobra built-in completion
> - Installation dans ~/.bashrc

**Estimated effort:** 2 heures

---

## 📝 How to use this TODO

**Pour chaque tâche:**
1. Copie le prompt complet
2. Colle dans Claude Code
3. Laisse Claude implémenter
4. Review + test
5. Commit
6. Passe à la tâche suivante

**Ordre recommandé:**
- P0 (#1-3) → Stabilité et fonctionnalités de base
- P1 (#4-7) → Performance et qualité
- P2 (#8-11) → Fonctionnalités avancées
- P3 (#12-15) → Nice-to-have

---

**Last updated:** 2026-02-09
**Total tasks:** 15
**Estimated total effort:** 3-4 semaines
