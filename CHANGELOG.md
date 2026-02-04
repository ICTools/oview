# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.0] - 2026-02-04

### BREAKING CHANGES

- **Suppression de l'intégration N8N** : N8N n'est plus installé comme infrastructure partagée
  - `oview install` ne crée plus de conteneur N8N
  - La configuration ne contient plus de paramètres N8N
  - Les conteneurs/volumes N8N existants doivent être supprimés manuellement
  - Voir [MIGRATION_0.2.0.md](MIGRATION_0.2.0.md) pour le guide de migration

### Changed

- Infrastructure simplifiée : Postgres+pgvector uniquement
- Réduction de l'utilisation des ressources (~50-200 MB économisés)
- Messages de sortie mis à jour pour refléter la nouvelle architecture
- Documentation mise à jour pour retirer toutes les références à N8N

### Removed

- Fonction `CreateN8nContainer()` supprimée de `internal/docker/container.go`
- Champs de configuration N8N supprimés de `GlobalConfig`
- Validation des paramètres N8N supprimée
- Affichage de l'URL N8N dans les commandes supprimé

## [0.1.0] - 2026-01-XX

### Added

- Commande `install` pour installer l'infrastructure globale (Postgres+pgvector, N8N)
- Commande `init` pour initialiser un nouveau projet
- Commande `up` pour démarrer le runtime d'un projet
- Commande `index` pour indexer le code source
- Commande `search` pour rechercher dans le code indexé
- Commande `benchmark` pour évaluer les performances de recherche
- Commande `verify` pour vérifier les embeddings
- Commande `monitor` pour surveillance en temps réel
- Serveur MCP pour intégration avec Claude Code
- Support multi-projets avec base de données isolée par projet
- Gestion automatique des ports Docker
- Configuration interactive pour l'initialisation
- Documentation complète (README, guides d'installation, guides techniques)

### Infrastructure

- Postgres avec extension pgvector
- Réseau Docker partagé pour tous les projets
- Volumes Docker persistants pour les données
- Configuration globale dans `~/.oview/config.yaml`
- Configuration par projet dans `.oview/config.yaml`
