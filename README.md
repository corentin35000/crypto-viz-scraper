# Crypto Viz - Scraper

## 🛠 Tech Stack

- Go (Language)
- Colly (Library)
- CI / CD (Github Actions)
- DockerCompose (Development-Local)
- Kubernetes (Development-Remote, Staging and Production)

<br /><br /><br /><br />

## 🏗️ Architecture de l'Application

Ce projet est structuré en quatre dépôts distincts, chacun jouant un rôle essentiel dans la collecte, le traitement et la visualisation des données. En suivant une architecture microservices, chaque dépôt est un composant indépendant, permettant une flexibilité et une maintenabilité optimales. Cette architecture modulaire permet également des mises à jour, des tests et des déploiements continus sans interférer avec les autres parties de l'application.

<br />

### Schéma d’Architecture

![Schéma d’architecture du projet Crypto Viz](crypto-viz-architecture.png)

<br />

### Composants et Responsabilités

#### 1. Frontend : `crypto-viz-frontend`
- **Rôle** : Fournit l'interface utilisateur (UI) pour visualiser en temps réel les données et analyses relatives aux cryptomonnaies.
- **Technologies** : Développé avec **Nuxt** et **TypeScript**, il intègre des bibliothèques de visualisation comme **D3.js** ou **Chart.js** pour afficher les données en prenant en compte leur dimension temporelle.
- **Responsabilité** : Le frontend consomme deux sources de données :
  - Les API du backend pour obtenir les données de marché et les analyses.
  - Le broker de messages pour recevoir les actualités filtrées du backend en temps réel.

#### 2. Scraper : `crypto-viz-scraper`
- **Rôle** : Ce composant se charge de collecter en temps réel les informations depuis un flux d'actualités sur les cryptomonnaies.
- **Technologies** : Construit en **Go** avec la bibliothèque **Colly**, qui permet de gérer efficacement un nombre important de requêtes simultanées grâce au parallélisme et à la gestion de la mémoire, en offrant de meilleures performances pour les scrapers intensifs.
- **Responsabilité** : Ce service fonctionne en continu et publie les actualités recueillies sur le broker de messages (`crypto-viz-broker`). En tant que producteur dans le modèle producteur/consommateur, il assure un flux constant de données.

#### 3. Backend : `crypto-viz-backend`
- **Rôle** : Analyse les données collectées par le scraper, filtre les informations pertinentes et les transforme pour les rendre exploitables par le frontend.
- **Technologies** : Construit avec **AdonisJS**, un framework backend offrant une structure complète pour la gestion des APIs et la logique métier.
- **Responsabilité** : 
  - Consomme les données via le broker de messages, filtre et enrichit les actualités pour créer des insights exploitables.
  - Publie les actualités filtrées sur un sujet dédié pour le frontend.
  - Expose des API HTTP permettant de récupérer les informations de marché, pour lesquelles il interroge directement des APIs externes de données financières.

#### 4. Broker : `crypto-viz-broker`
- **Rôle** : Assure la communication entre le scraper, le backend, et le frontend en utilisant un système de messages.
- **Technologies** : Utilise **NATS** comme broker de messages, permettant une gestion performante et asynchrone des flux de données.
- **Responsabilité** : 
  - Gère la distribution des messages entre les services en maintenant un flux de données continu entre les composants.
  - Permet au scraper de publier des messages de manière indépendante, que le backend peut consommer et analyser.
  - Facilite la scalabilité en permettant d'ajouter des instances supplémentaires de chaque composant sans créer de dépendances directes.

<br />

### Déroulement des Données et Interactions

1. Le **scraper** publie les dernières actualités sur le sujet `crypto.news` du broker.
2. Le **backend** souscrit au sujet `crypto.news`, filtre et traite les informations, puis publie les actualités filtrées et enrichies sur un nouveau sujet `crypto.news.filtered`.
3. Le **frontend** souscrit au sujet `crypto.news.filtered` pour obtenir en temps réel les informations actualisées.
4. Pour les données de marché, le **frontend** interroge directement le backend via des appels HTTP, qui utilise des APIs externes pour récupérer les informations les plus récentes.

<br />

### Avantages de cette Architecture

- **Indépendance des Composants** : Chaque service est autonome et peut être mis à jour, redémarré ou mis à l’échelle indépendamment.
- **Scalabilité** : Le broker de messages permet d’ajouter des instances supplémentaires de chaque service pour gérer une charge accrue sans modification de l'architecture.
- **Temps Réel** : Le modèle producteur/consommateur via NATS permet des communications en temps réel, assurant que les utilisateurs disposent toujours des données les plus récentes.
- **Flexibilité** : La séparation des responsabilités permet d'ajuster ou de remplacer un composant sans perturber le reste du système.

<br /><br /><br /><br />

## 📁 Organisation du Code Source du Scraper

crypto-viz-scraper/ <br />
├── Dockerfile            `# Instructions pour construire l'image Docker` <br />
├── docker-compose.yml    `# Configuration Docker Compose pour lancer l'application` <br />
├── .air.toml             `# Configuration d'Air pour le rechargement automatique en développement` <br />
├── go.mod                `# Dépendances Go du projet` <br />
├── go.sum                `# Vérification d'intégrité des dépendances Go` <br />
└── src/                  `# Code source de l'application` <br />
    ├── main.go           `# Point d'entrée de l'application` <br />
    ├── collector.go      `# Configuration et gestion du collecteur Colly` <br />
    ├── processor.go      `# Logique du scraper` <br />
    ├── nats.go           `# Service Nats pour le broker de message` <br />

<br /><br /><br /><br />

## 📦 Versionning

On utilise la convention SemVer : https://semver.org/lang/fr/ <br /><br />
Pour une Release classique : MAJOR.MINOR.PATCH <br />
Pour une Pre-Release, exemples : MAJOR.MINOR.PATCH-rc.0 OR MAJOR.MINOR.PATCH-beta.3 <br /><br />

Nous utilison release-please de Google pour versionner, via Github Actions. <br />
Pour que cela sois pris en compte il faut utiliser les conventionnal commits : https://www.conventionalcommits.org/en/v1.0.0/ <br />
Release Please crée une demande d'extraction de version après avoir remarqué que la branche par défaut contient des « unités publiables » depuis la dernière version. Une unité publiable est un commit sur la branche avec l'un des préfixes suivants : `feat` / `feat!` et `fix` / `fix!`. <br /><br />

La première Release que créer release-please automatiquement est la version : 1.0.0 <br />
Pour créer une Pre-Release faire un commit vide, par exemple si on'ai à la version 1.0.0, on peut faire :

```bash
git commit --allow-empty -m "chore: release 1.1.0-rc.0" -m "Release-As: 1.1.0-rc.0"
```

<br /><br /><br /><br />

## 🚀 Conventions de Commit

Nous utilisons les conventions de commit pour maintenir une cohérence dans l'historique du code et faciliter le versionnement automatique avec release-please. Voici les types de commits que nous utilisons, ainsi que leur impact sur le versionnage :

- feat : Introduction d'une nouvelle fonctionnalité pour l'utilisateur. Entraîne une augmentation de la version mineure (par exemple, de 1.0.0 à 1.1.0).

- feat! : Introduction d'une nouvelle fonctionnalité avec des modifications incompatibles avec les versions antérieures (breaking changes). Entraîne une augmentation de la version majeure (par exemple, de 1.0.0 à 2.0.0).

- fix : Correction d'un bug pour l'utilisateur. Entraîne une augmentation de la version patch (par exemple, de 1.0.0 à 1.0.1).

- fix! : Correction d'un bug avec des modifications incompatibles avec les versions antérieures (breaking changes). Entraîne une augmentation de la version majeure.

- docs : Changements concernant uniquement la documentation. N'affecte pas la version.

- style : Changements qui n'affectent pas le sens du code (espaces blancs, mise en forme, etc.). N'affecte pas la version.

- refactor : Modifications du code qui n'apportent ni nouvelle fonctionnalité ni correction de bug. N'affecte pas la version.

- perf : Changements de code qui améliorent les performances. Peut entraîner une augmentation de la version mineure.

- test : Ajout ou correction de tests. N'affecte pas la version.

- chore : Changements qui ne modifient ni les fichiers source ni les tests (par exemple, mise à jour des dépendances). N'affecte pas la version.

- ci : Changements dans les fichiers de configuration et les scripts d'intégration continue (par exemple, GitHub Actions). N'affecte pas la version.

- build : Changements qui affectent le système de build ou les dépendances externes (par exemple, npm, Docker). N'affecte pas la version.

- revert : Annulation d'un commit précédent. N'affecte pas la version.

Pour indiquer qu'un commit introduit des modifications incompatibles avec les versions antérieures (breaking changes), ajoutez un ! après le type de commit, par exemple feat! ou fix!.

Pour plus de détails sur les conventions de commit, consultez : [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/)

<br /><br /><br /><br />

## 📚 Domains of different environments

- Production : https://test.crzcommon.com
- Staging : https://staging.test.crzcommon.com
- Development-Remote : https://dev.test.crzcommon.com

<br /><br /><br /><br />

## ⚙️ Setup Environment Development

1. Clone the project repository using the following commands :

```bash
git clone git@github.com:corentin35000/crypto-viz-scraper.git
```

2. Steps by Platform :

```bash
# Windows :
1. Requirements : Windows >= 10
2. Download and Install WSL2 : https://learn.microsoft.com/fr-fr/windows/wsl/install
3. Download and Install Docker Desktop : https://docs.docker.com/desktop/install/windows-install/

# macOS :
1. Requirements : macOS Intel x86_64 or macOS Apple Silicon arm64
2. Requirements (2) : macOS 11.0 (Big Sur)
2. Download and Install Docker Desktop : https://docs.docker.com/desktop/install/mac-install/

# Linux (Ubuntu / Debian) :
1. Requirements : Ubuntu >= 20.04 or Debian >= 10
2. Download and Install Docker (Ubuntu) : https://docs.docker.com/engine/install/ubuntu/
3. Download and Install Docker (Debian) : https://docs.docker.com/engine/install/debian/
```

<br /><br /><br /><br />

## 🔄 Cycle Development

1. macOS / Windows : Open Docker Desktop
2. Run command :
```bash
   docker-compose up
```

<br /><br /><br /><br />

## 🚀 Production

### ⚙️➡️ Automatic Distribution Process (CI / CD)

#### Si c'est un nouveau projet suivez les instructions :

1. Ajoutées les SECRETS_GITHUB pour :
   - DOCKER_HUB_USERNAME
   - DOCKER_HUB_ACCESS_TOKEN
   - KUBECONFIG
   - PAT (crée un nouveau token si besoin sur le site de github puis dans le menu du "Profil" puis -> "Settings" -> "Developper Settings' -> 'Personnal Access Tokens' -> Tokens (classic))
