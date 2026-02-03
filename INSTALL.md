# Guide d'Installation oview

## Méthode 1 : Script d'installation automatique (Recommandé)

### Installation rapide

```bash
cd /path/to/oview
./install.sh
```

Ou depuis le dépôt (quand publié) :

```bash
curl -fsSL https://raw.githubusercontent.com/yourusername/oview/main/install.sh | bash
```

### Ce que fait le script

1. ✅ Vérifie les prérequis (OS, Docker)
2. ✅ Propose de télécharger un binaire ou compiler
3. ✅ Installe dans `/usr/local/bin/`
4. ✅ Vérifie l'installation
5. ✅ Propose de lancer `oview install` (infrastructure)
6. ✅ Affiche les prochaines étapes

### Options du script

```bash
# Installation normale (interactive)
./install.sh

# Désinstallation
./install.sh uninstall
```

## Méthode 2 : Installation manuelle

### Prérequis

- **Docker** : Obligatoire
- **Go 1.23+** : Uniquement pour compiler depuis les sources

#### Installer Docker

**Ubuntu/Debian :**
```bash
sudo apt-get update
sudo apt-get install docker.io
sudo systemctl start docker
sudo systemctl enable docker
sudo usermod -aG docker $USER
newgrp docker
```

**macOS :**
```bash
brew install --cask docker
# Ou télécharger Docker Desktop
```

**Fedora :**
```bash
sudo dnf install docker
sudo systemctl start docker
sudo systemctl enable docker
```

#### Installer Go (optionnel)

```bash
# Via gestionnaire de paquets
# Ubuntu/Debian
sudo apt-get install golang-go

# macOS
brew install go

# Ou télécharger depuis https://go.dev/dl/
```

### Depuis les sources

```bash
# 1. Cloner le dépôt
git clone https://github.com/yourusername/oview.git
cd oview

# 2. Compiler
go build -o oview .

# 3. Installer
sudo cp oview /usr/local/bin/oview
sudo chmod +x /usr/local/bin/oview

# 4. Vérifier
oview version

# 5. Installer l'infrastructure
oview install
```

### Depuis un binaire précompilé

```bash
# 1. Télécharger le binaire pour votre plateforme
# Linux AMD64
wget https://github.com/yourusername/oview/releases/latest/download/oview-linux-amd64

# macOS ARM64 (M1/M2)
wget https://github.com/yourusername/oview/releases/latest/download/oview-darwin-arm64

# 2. Renommer et installer
mv oview-* oview
chmod +x oview
sudo mv oview /usr/local/bin/

# 3. Vérifier
oview version

# 4. Installer l'infrastructure
oview install
```

## Méthode 3 : Installation locale (sans sudo)

Si vous n'avez pas les droits admin :

```bash
# 1. Créer un répertoire bin dans votre home
mkdir -p ~/bin

# 2. Ajouter à votre PATH (une seule fois)
echo 'export PATH="$HOME/bin:$PATH"' >> ~/.bashrc  # ou ~/.zshrc
source ~/.bashrc  # ou ~/.zshrc

# 3. Compiler et copier
go build -o oview .
cp oview ~/bin/

# 4. Utiliser
oview version
oview install
```

## Vérification de l'installation

### Vérifier le binaire

```bash
# Commande disponible ?
which oview
# /usr/local/bin/oview

# Version ?
oview version
# oview version 0.1.0

# Aide ?
oview --help
```

### Vérifier Docker

```bash
# Docker fonctionne ?
docker ps

# Conteneurs oview créés ?
docker ps | grep oview
# oview-postgres
# oview-n8n
```

### Vérifier la config

```bash
# Config globale existe ?
ls -la ~/.oview/
# config.yaml

# Contenu de la config ?
cat ~/.oview/config.yaml
```

## Workflow complet d'installation

### Installation zéro à héro (5 minutes)

```bash
# 1. Prérequis Docker (si pas déjà fait)
# Suivre les instructions ci-dessus pour votre OS

# 2. Installation oview
cd /path/to/oview
./install.sh
# Répondre aux questions interactives

# 3. Vérification
oview version

# 4. Test sur un projet
cd ~/Documents/mon-projet
oview init
# Configuration interactive

# 5. Setup runtime
oview up

# 6. Indexation
oview index

# ✅ Prêt à l'emploi !
```

## Mise à jour

### Avec le script d'installation

```bash
# 1. Désinstaller l'ancienne version (garde les données)
oview uninstall --keep-data --keep-config

# 2. Mettre à jour les sources
cd /path/to/oview
git pull

# 3. Réinstaller
./install.sh

# 4. Vérifier
oview version
```

### Manuellement

```bash
# 1. Désinstaller
oview uninstall --keep-data --keep-config

# 2. Recompiler
cd /path/to/oview
git pull
go build -o oview .

# 3. Réinstaller
sudo cp oview /usr/local/bin/oview

# 4. Réinstaller l'infrastructure
oview install
# Reconnecte aux volumes existants
```

## Désinstallation

### Avec le script

```bash
./install.sh uninstall
```

Le script demande :
- Supprimer l'infrastructure Docker ?
- Supprimer la configuration ?

### Avec oview

```bash
# Désinstallation complète
oview uninstall
sudo rm /usr/local/bin/oview

# Ou garder les données
oview uninstall --keep-data --keep-config
sudo rm /usr/local/bin/oview
```

### Manuelle (complète)

```bash
# 1. Infrastructure Docker
docker stop oview-postgres oview-n8n
docker rm oview-postgres oview-n8n
docker volume rm oview-postgres-data oview-n8n-data
docker network rm oview-net

# 2. Binaire
sudo rm /usr/local/bin/oview

# 3. Configuration
rm -rf ~/.oview

# 4. Projets (optionnel)
find ~ -type d -name ".oview" -exec rm -rf {} +
```

## Dépannage

### "Docker is not running"

**Linux :**
```bash
sudo systemctl start docker
sudo systemctl status docker
```

**macOS :**
```bash
# Lancer Docker Desktop depuis Applications
```

### "Permission denied" lors de docker ps

```bash
# Ajouter votre user au groupe docker
sudo usermod -aG docker $USER
newgrp docker

# Ou redémarrer votre session
```

### "command not found: oview"

```bash
# Vérifier l'installation
ls -la /usr/local/bin/oview

# Vérifier le PATH
echo $PATH | grep /usr/local/bin

# Si pas dans PATH, ajouter :
echo 'export PATH="/usr/local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

### Build échoue avec "Go version too old"

```bash
# Vérifier la version
go version

# Mettre à jour Go
# Ubuntu (via snap)
sudo snap install go --classic

# macOS
brew upgrade go

# Ou télécharger depuis https://go.dev/dl/
```

### "Failed to download binary"

Le script essaie de télécharger un binaire précompilé. Si ça échoue :

1. Vérifier votre connexion Internet
2. Le script bascule automatiquement sur la compilation
3. Ou compiler manuellement : `go build -o oview .`

### Installation réussit mais "oview install" échoue

Vérifier Docker :
```bash
# Docker tourne ?
docker ps

# Ports disponibles ?
sudo lsof -i :5432  # Postgres
sudo lsof -i :5678  # n8n

# Si ports occupés, ils seront automatiquement changés
```

## Installation pour le développement

Si vous comptez développer sur oview :

```bash
# 1. Fork et clone
git clone https://github.com/yourfork/oview.git
cd oview

# 2. Installer les dépendances
go mod download

# 3. Compiler en mode dev
go build -o oview .

# 4. Lancer depuis le dossier actuel
./oview version

# 5. Ou créer un lien symbolique
sudo ln -sf $(pwd)/oview /usr/local/bin/oview

# Maintenant vous pouvez recompiler et tester facilement
go build -o oview . && oview version
```

## Installation en production

Pour un serveur de production :

```bash
# 1. Télécharger le binaire
wget https://github.com/yourusername/oview/releases/latest/download/oview-linux-amd64
mv oview-linux-amd64 oview
chmod +x oview
sudo mv oview /usr/local/bin/

# 2. Créer un utilisateur dédié
sudo useradd -r -s /bin/false oview
sudo usermod -aG docker oview

# 3. Configurer systemd (optionnel)
sudo tee /etc/systemd/system/oview.service > /dev/null <<EOF
[Unit]
Description=oview Infrastructure
After=docker.service
Requires=docker.service

[Service]
Type=oneshot
User=oview
ExecStart=/usr/local/bin/oview install
RemainAfterExit=yes

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl enable oview
sudo systemctl start oview

# 4. Vérifier
docker ps | grep oview
```

## Installation multi-utilisateurs

Pour un serveur partagé :

```bash
# Chaque utilisateur peut avoir ses propres projets
# L'infrastructure Docker est partagée

# User 1
su - user1
oview init  # Dans son projet
oview up    # Crée sa DB

# User 2
su - user2
oview init  # Dans son projet
oview up    # Crée sa DB

# Les deux utilisent le même Postgres/n8n
docker ps | grep oview
# oview-postgres (partagé)
# oview-n8n (partagé)

# Mais chacun a sa propre DB
docker exec oview-postgres psql -U oview -l
# oview_user1_project
# oview_user2_project
```

## Plateformes supportées

| OS | Architecture | Status |
|----|--------------|--------|
| Linux | AMD64 | ✅ Testé |
| Linux | ARM64 | ✅ Supporté |
| macOS | AMD64 | ✅ Supporté |
| macOS | ARM64 (M1/M2) | ✅ Supporté |
| Windows | AMD64 | ⚠️ Expérimental (WSL2) |

## Support

- 📖 Documentation : [README.md](README.md)
- 🐛 Issues : https://github.com/yourusername/oview/issues
- 💬 Discussions : https://github.com/yourusername/oview/discussions

---

**Installation réussie ? Lancez `oview init` dans votre premier projet !** 🚀
