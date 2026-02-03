# Guide de Désinstallation oview

## Commande de base

```bash
oview uninstall
```

Cette commande supprime **toute l'infrastructure globale** oview.

## Ce qui est supprimé

### Par défaut (suppression complète)

```bash
oview uninstall
```

**Supprime :**
- ✅ Conteneur `oview-postgres`
- ✅ Conteneur `oview-n8n`
- ✅ Volume `oview-postgres-data` (⚠️ **TOUTES les bases de données**)
- ✅ Volume `oview-n8n-data` (⚠️ **TOUS les workflows**)
- ✅ Réseau `oview-net`
- ✅ Fichier `~/.oview/config.yaml`

**Résultat :** Nettoyage complet, comme si oview n'avait jamais été installé.

## Options

### 1. Garder les données (réinstallation future)

```bash
oview uninstall --keep-data
```

**Préserve :**
- 💾 Volume Postgres (toutes vos bases de données)
- 💾 Volume n8n (tous vos workflows)

**Supprime :**
- 🐳 Conteneurs
- 🌐 Réseau
- 📄 Config

**Cas d'usage :**
- Mettre à jour oview (désinstaller/réinstaller)
- Libérer de la RAM sans perdre les données
- Tester une nouvelle version

**Réinstallation :**
```bash
oview install  # Reconnecte aux volumes existants
```

### 2. Garder la configuration

```bash
oview uninstall --keep-config
```

**Préserve :**
- 📄 `~/.oview/config.yaml` (ports, mots de passe, etc.)

**Cas d'usage :**
- Garder la config pour une réinstallation rapide
- Éviter les conflits de ports lors de la réinstallation

### 3. Mode force (pas de confirmation)

```bash
oview uninstall --force
```

**Attention :** Supprime immédiatement sans demander de confirmation.

**Cas d'usage :**
- Scripts automatisés
- CI/CD pipelines

### 4. Combinaisons

```bash
# Garde data + config
oview uninstall --keep-data --keep-config

# Force + garde data
oview uninstall -f --keep-data
```

## Workflows courants

### Désinstallation complète (reset total)

```bash
# 1. Désinstaller oview
oview uninstall

# 2. Supprimer le binaire
sudo rm /usr/local/bin/oview

# 3. Nettoyer les projets (optionnel)
find ~ -type d -name ".oview" -exec rm -rf {} +
```

**Résultat :** Plus aucune trace d'oview sur votre système.

### Mise à jour d'oview

```bash
# 1. Désinstaller en gardant les données
oview uninstall --keep-data --keep-config

# 2. Mettre à jour le binaire
cd /path/to/oview
git pull
go build -o oview .
sudo cp oview /usr/local/bin/oview

# 3. Réinstaller (reconnecte aux données existantes)
oview install
```

**Résultat :** oview mis à jour, données préservées.

### Libérer de la RAM temporairement

```bash
# Stopper sans supprimer les données
oview uninstall --keep-data --keep-config
```

Plus tard :
```bash
oview install  # Redémarre tout
```

### Migration vers un autre système

**Sur l'ancienne machine :**
```bash
# 1. Sauvegarder les volumes
docker run --rm -v oview-postgres-data:/data -v $(pwd):/backup \
  ubuntu tar czf /backup/oview-postgres-backup.tar.gz -C /data .

docker run --rm -v oview-n8n-data:/data -v $(pwd):/backup \
  ubuntu tar czf /backup/oview-n8n-backup.tar.gz -C /data .

# 2. Copier les backups vers la nouvelle machine
scp oview-*.tar.gz user@new-machine:~/
```

**Sur la nouvelle machine :**
```bash
# 1. Créer les volumes
docker volume create oview-postgres-data
docker volume create oview-n8n-data

# 2. Restaurer
docker run --rm -v oview-postgres-data:/data -v $(pwd):/backup \
  ubuntu tar xzf /backup/oview-postgres-backup.tar.gz -C /data

docker run --rm -v oview-n8n-data:/data -v $(pwd):/backup \
  ubuntu tar xzf /backup/oview-n8n-backup.tar.gz -C /data

# 3. Installer oview
oview install
```

## Vérification

### Avant désinstallation

```bash
# Lister ce qui sera supprimé
docker ps -a --filter "name=oview"
docker volume ls --filter "name=oview"
docker network ls --filter "name=oview"
ls -la ~/.oview/
```

### Après désinstallation

```bash
# Vérifier que tout est supprimé
docker ps -a --filter "name=oview"      # Devrait être vide
docker volume ls --filter "name=oview"  # Devrait être vide (sauf si --keep-data)
docker network ls --filter "name=oview" # Devrait être vide
ls -la ~/.oview/                        # Devrait ne pas exister (sauf si --keep-config)
```

## Suppression manuelle (si problème)

Si `oview uninstall` échoue, nettoyage manuel :

```bash
# 1. Arrêter et supprimer les conteneurs
docker stop oview-postgres oview-n8n
docker rm oview-postgres oview-n8n

# 2. Supprimer les volumes
docker volume rm oview-postgres-data oview-n8n-data

# 3. Supprimer le réseau
docker network rm oview-net

# 4. Supprimer la config
rm -rf ~/.oview

# 5. Supprimer le binaire
sudo rm /usr/local/bin/oview
```

## Récupération après désinstallation accidentelle

### Si vous avez utilisé --keep-data

```bash
# Les volumes existent toujours
docker volume ls | grep oview

# Réinstaller simplement
oview install

# Vos projets sont toujours là !
cd ~/Documents/chapitreneuf
oview index  # Reconnecte à la base existante
```

### Si vous n'avez PAS utilisé --keep-data

**Les données sont perdues.** Vous devez :

1. Réinstaller : `oview install`
2. Réindexer chaque projet : `cd project && oview index`

**Conseil :** Toujours utiliser `--keep-data` sauf si vous êtes sûr de vouloir tout supprimer.

## Suppression des projets

`oview uninstall` ne touche **PAS** aux dossiers `.oview/` dans vos projets.

Pour nettoyer :

```bash
# Supprimer .oview/ d'un projet spécifique
cd ~/Documents/chapitreneuf
rm -rf .oview

# Supprimer tous les .oview/ (attention!)
find ~ -type d -name ".oview" -exec rm -rf {} +
```

## FAQ

### Puis-je désinstaller pendant que des projets sont en cours ?

Oui, mais :
- Les connexions DB seront coupées
- Sauvegardez votre travail d'abord
- Les conteneurs s'arrêtent proprement

### Est-ce que ça supprime mes fichiers de code ?

**Non.** Seuls les dossiers `.oview/` et l'infrastructure Docker sont concernés.
Votre code source n'est jamais touché.

### Combien d'espace disque je récupère ?

Environ :
- Volumes Postgres : 100-500 MB (dépend du nombre de projets indexés)
- Volume n8n : 50-200 MB
- Conteneurs : 500 MB
- **Total : ~1-2 GB**

### Puis-je désinstaller sans le binaire oview ?

Oui, nettoyage manuel :
```bash
docker rm -f oview-postgres oview-n8n
docker volume rm oview-postgres-data oview-n8n-data
docker network rm oview-net
rm -rf ~/.oview
```

### La désinstallation nécessite-t-elle sudo ?

**Non** pour la désinstallation Docker (elle utilise votre accès Docker normal).

**Oui** uniquement pour supprimer le binaire :
```bash
sudo rm /usr/local/bin/oview
```

## Réinstallation après désinstallation complète

```bash
# 1. Réinstaller l'infrastructure
oview install

# 2. Pour chaque projet, réinitialiser
cd ~/Documents/mon-projet
oview init --force    # Regénère .oview/
oview up              # Recrée la DB
oview index           # Réindexe le code
```

**Durée estimée :** 5-10 minutes par projet.

## Scénarios d'urgence

### "J'ai lancé uninstall par erreur, CTRL+C fonctionne ?"

**Oui**, si vous interrompez pendant la confirmation.

**Non**, si vous avez déjà confirmé. Dans ce cas :
- Si `--keep-data` : vos données sont safe
- Sinon : les conteneurs arrêtés peuvent encore être redémarrés pendant quelques secondes

### "J'ai supprimé par erreur sans --keep-data"

**Si c'est très récent (< 1 minute) :**

1. Ne lancez PAS `docker volume prune`
2. Les volumes peuvent encore exister temporairement
3. Vérifiez : `docker volume ls`
4. Si présents, réinstallez vite : `oview install`

**Sinon :** Données perdues, il faut réindexer.

---

**En résumé :**
- `oview uninstall` : suppression interactive et sûre
- `--keep-data` : votre filet de sécurité
- `--force` : pour les scripts
- Pas de sudo nécessaire (sauf pour supprimer le binaire)
