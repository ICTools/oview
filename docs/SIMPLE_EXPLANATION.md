# 🎓 Explication simple : Comment ça marche ?

## En une phrase

**oview transforme votre code en "vecteurs intelligents" stockés dans une base de données, permettant à Claude de chercher par "sens" au lieu de mots-clés.**

---

## 🍰 L'analogie du gâteau

Imaginez votre codebase comme un **livre de recettes géant**.

### Sans RAG (recherche classique)

```
Vous: "Comment faire un gâteau au chocolat ?"
Recherche: Ctrl+F "gâteau au chocolat"

Résultat:
❌ Rate la recette "Cake au cacao"
❌ Rate la recette "Fondant chocolaté"
✅ Trouve "Gâteau au chocolat" (si écrit exactement comme ça)
```

### Avec RAG (oview)

```
Vous: "Comment faire un gâteau au chocolat ?"

oview:
1. Comprend que vous cherchez quelque chose lié à:
   - Dessert chocolaté
   - Cuisson
   - Pâtisserie

2. Trouve automatiquement:
   ✅ "Cake au cacao"
   ✅ "Fondant chocolaté"
   ✅ "Gâteau au chocolat"
   ✅ "Brownie"
   ✅ "Moelleux au chocolat"

3. Classe par pertinence (92%, 88%, 76%...)
```

**Magie :** RAG comprend le **SENS**, pas juste les mots !

---

## 🔢 Comment ça marche techniquement ?

### Étape 1 : Transformer en nombres (Embeddings)

Chaque morceau de code devient un vecteur (liste de nombres) :

```
Code:
class UserController {
    public function login() { ... }
}

Devient:
[0.23, -0.45, 0.12, 0.78, -0.34, 0.56, ... ] (768 nombres)
        ↑
    Ces nombres capturent le SENS du code
```

### Étape 2 : Stocker dans une base de données

```
PostgreSQL + pgvector:
┌────┬─────────────────┬─────────────────┬─────────────────┐
│ ID │ Fichier         │ Code            │ Vecteur         │
├────┼─────────────────┼─────────────────┼─────────────────┤
│ 1  │ UserControl...  │ login() {...}   │ [0.23, -0.45...]│
│ 2  │ AuthService...  │ authenticate()  │ [0.21, -0.48...]│
│ 3  │ Security.yaml   │ security: ...   │ [0.19, -0.52...]│
└────┴─────────────────┴─────────────────┴─────────────────┘
```

### Étape 3 : Chercher par similarité

Quand vous cherchez "authentication" :

1. **Transformer votre question en vecteur**
   ```
   "authentication" → [0.21, -0.48, 0.11, ...]
   ```

2. **Comparer avec tous les vecteurs stockés**
   ```
   [0.21, -0.48, ...] vs [0.23, -0.45, ...] = 92% similaire ✅
   [0.21, -0.48, ...] vs [0.89, 0.34, ...]  = 23% similaire ❌
   ```

3. **Retourner les plus similaires**
   ```
   1. login() - 92%
   2. authenticate() - 88%
   3. security config - 76%
   ```

---

## 🎬 Un exemple concret

Vous demandez à Claude : **"Où est le code qui gère les erreurs ?"**

### Ce qui se passe en coulisses :

```
1. Vous → Claude Code
   "Où est le code qui gère les erreurs ?"

2. Claude → oview MCP
   search("error handling code")

3. oview → Ollama
   Transforme "error handling code" en vecteur
   [0.34, -0.67, 0.23, ...]

4. oview → PostgreSQL
   SELECT * FROM chunks
   ORDER BY similarity_to([0.34, -0.67, 0.23, ...])
   LIMIT 5

5. PostgreSQL → oview
   Trouve:
   - ExceptionHandler.php (94% similaire)
   - ErrorController.php (89% similaire)
   - LoggerService.php (81% similaire)

6. oview → Claude
   Voici les 3 chunks les plus pertinents...

7. Claude → Vous
   "Le code de gestion d'erreurs se trouve dans:
    - ExceptionHandler.php pour les exceptions
    - ErrorController.php pour les pages d'erreur
    - LoggerService.php pour le logging"
```

**Temps total : ~1 seconde**

---

## 💡 Pourquoi c'est puissant ?

### 1. Compréhension sémantique

```
Votre recherche: "comment les utilisateurs se connectent"

RAG trouve automatiquement:
✅ login()
✅ authenticate()
✅ signin()
✅ verifyCredentials()
✅ UserController
✅ AuthService
✅ security.yaml

Même si ces mots ne sont PAS dans votre recherche !
```

### 2. Fonctionne en plusieurs langues

```
Recherche en français: "gestion des erreurs"
Trouve du code en anglais: "error handling", "exception", "try/catch"
```

### 3. Trouve le code similaire

```
"Je veux faire quelque chose comme dans UserController"
→ Trouve tous les controllers similaires
```

---

## 🎯 Les 3 cas d'usage principaux

### 1️⃣ Explorer un nouveau projet

```
Vous: "Comment marche le système de cache ici ?"
Claude: [Cherche dans le RAG]
        "Le système utilise Redis avec un CacheManager..."
```

### 2️⃣ Modifier du code en sécurité

```
Vous: "Je veux modifier UserController"
Claude: [Récupère le contexte via RAG]
        [Voit que UserController dépend de AuthService]
        "Attention, cette classe est utilisée par AuthService..."
```

### 3️⃣ Trouver des exemples

```
Vous: "Montre-moi un exemple d'API REST dans ce projet"
Claude: [Cherche dans le RAG]
        "Voici 3 exemples d'API REST:
         1. UserApiController
         2. ProductApiController
         ..."
```

---

## 🔐 C'est sûr ?

**Oui ! Tout reste sur votre machine :**

```
┌─────────────────────────────────────────┐
│         Votre ordinateur                │
│                                         │
│  ┌─────────────┐    ┌──────────────┐   │
│  │ Votre code  │───►│    Ollama    │   │
│  │             │    │   (local)    │   │
│  └─────────────┘    └──────┬───────┘   │
│                             │           │
│                             ▼           │
│                    ┌──────────────┐     │
│                    │  PostgreSQL  │     │
│                    │   (Docker)   │     │
│                    └──────────────┘     │
│                                         │
│  🔒 Rien ne sort de votre machine !    │
└─────────────────────────────────────────┘
```

**Exception :** Si vous utilisez OpenAI au lieu d'Ollama
- Les morceaux de code sont envoyés à OpenAI pour générer les embeddings
- Mais **jamais** le code complet, juste des petits chunks

---

## ⚡ C'est rapide ?

**Oui !**

```
Indexation de 500 fichiers:   ~5 secondes
Recherche dans 1000 chunks:    ~200 millisecondes
Total pour une question:       ~1 seconde
```

**Comparaison :**
- Lire manuellement 500 fichiers: **~30 minutes**
- Chercher avec grep: **~5 secondes** (mais beaucoup de bruit)
- Chercher avec RAG: **~1 seconde** (résultats pertinents)

---

## 🤔 Questions fréquentes

**Q: C'est comme ChatGPT qui lit mon code ?**
R: Non ! ChatGPT lit et "comprend" le code. oview le transforme juste en vecteurs pour chercher rapidement.

**Q: Ça remplace GitHub Copilot ?**
R: Non, c'est complémentaire. Copilot suggère du code, oview aide à **chercher** et **comprendre** le code existant.

**Q: Je dois tout ré-indexer à chaque modification ?**
R: Oui pour l'instant. Une indexation incrémentale est prévue dans la roadmap.

**Q: Ça marche avec n'importe quel langage ?**
R: Oui ! PHP, JavaScript, Python, Go, Rust... Les embeddings capturent le sens, pas la syntaxe.

**Q: Combien ça coûte ?**
R: **0€ avec Ollama (local)** ou ~0.02€ par million de tokens avec OpenAI.

**Q: C'est difficile à installer ?**
R: Non ! 3 commandes:
```bash
oview init
oview up
oview index
```

---

## 🎓 Pour aller plus loin

- **Guide technique complet**: `docs/HOW_IT_WORKS.md`
- **Installation MCP**: `docs/QUICK_START_MCP.md`
- **Configuration avancée**: `docs/MCP_INTEGRATION.md`

---

## 🎉 En résumé

**oview = Google pour votre code**

Au lieu de chercher par mots-clés, vous cherchez par **sens**.

```
Sans oview:  grep "cache"           → 247 résultats, 90% inutiles
Avec oview:  search "cache system"  → 5 résultats, 95% pertinents
```

**Et Claude Code peut utiliser ça automatiquement !**

```
Vous: "Comment marche le cache ?"
Claude: [Utilise oview automatiquement]
        [Trouve le code pertinent]
        [Vous explique]
```

C'est tout ! 🚀
