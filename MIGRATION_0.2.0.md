# Migration vers 0.2.0 : Suppression de N8N

## Changements

N8N n'est plus installé comme infrastructure partagée. Cette modification simplifie l'architecture d'oview en se concentrant uniquement sur les fonctionnalités de base : indexation de code et recherche sémantique via Postgres+pgvector.

## Si vous avez une installation existante avec N8N

### 1. Sauvegarder vos workflows (optionnel)

Si vous avez créé des workflows N8N que vous souhaitez conserver :

```bash
docker run --rm -v oview-n8n-data:/data -v $(pwd):/backup \
  ubuntu tar czf /backup/n8n-backup.tar.gz -C /data .
```

### 2. Nettoyer les ressources N8N

Avant de mettre à jour oview, supprimez manuellement les ressources N8N :

```bash
# Arrêter et supprimer le conteneur N8N
docker stop oview-n8n
docker rm oview-n8n

# Supprimer le volume N8N
docker volume rm oview-n8n-data
```

### 3. Mettre à jour oview

```bash
# Désinstaller en conservant les données et la config
oview uninstall --keep-data --keep-config

# Installer la nouvelle version
# (après avoir mis à jour le binaire oview)
oview install
```

## Pour les nouvelles installations

Aucune action spécifique requise. Installez simplement oview normalement :

```bash
oview install
```

## Alternatives pour l'orchestration

Si vous avez besoin d'orchestration de workflows pour automatiser vos processus :

- **GitHub Actions** : Pour CI/CD et automatisations basées sur des événements git
- **Jenkins** : Pour pipelines de build complexes
- **Airflow** : Pour orchestration de workflows de données
- **Scripts d'automation custom** : Pour des besoins spécifiques simples

## Impact sur les fonctionnalités

La suppression de N8N **n'affecte pas** les fonctionnalités principales d'oview :

- ✅ Indexation de code avec `oview index`
- ✅ Recherche sémantique via le serveur MCP
- ✅ Gestion de base de données Postgres+pgvector
- ✅ Support multi-projets

## Questions fréquentes

### Pourquoi supprimer N8N ?

N8N était installé dans les versions initiales mais n'était pas utilisé par les fonctionnalités implémentées. Sa suppression :
- Simplifie l'architecture
- Réduit l'utilisation des ressources (~50-200 MB)
- Facilite la maintenance

### Puis-je utiliser N8N séparément ?

Oui, vous pouvez installer N8N indépendamment si vous en avez besoin :

```bash
docker run -d --name my-n8n \
  -p 5678:5678 \
  -v n8n-data:/home/node/.n8n \
  n8nio/n8n:latest
```

### Mes anciennes configurations vont-elles fonctionner ?

Oui. Les anciennes configurations contenant des champs N8N seront chargées sans erreur. Les champs N8N inconnus seront simplement ignorés.
