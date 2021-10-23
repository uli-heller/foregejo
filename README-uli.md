Uli's Bau-Anleitung
===================

Build-Container einrichten
--------------------------

* Ausgangspunkt: Ubuntu-20.04 Basisinstallation
* Zusatzpakete installieren
    * `sudo apt install golang-go` ... installiert eine zu alte Version
    * `sudo apt install nodejs npm`
    * `sudo apt install make`
* Neuere Version von Go installieren
    * `mkdir Software`
    * `cd Software`
    * `GO_VERSION=1.17.1`
    * `wget https://golang.org/dl/go${GO_VERSION}.linux-amd64.tar.gz`
    * `rm -rf go`
    * `gzip -cd go${GO_VERSION}.linux-amd64.tar.gz|tar xf -`
* PATH erweitern in ~/.bashrc:
    ```diff
    +PATH="${HOME}/Software/go/bin:${PATH}"
    +export PATH
    ```
* Neuere NodeJS-Version installieren
    * `mkdir Software`
    * `cd Software`
    * `NODEJS_VERSION=14.18.0`
    * `wget https://nodejs.org/dist/v${NODEJS_VERSION}/node-v${NODEJS_VERSION}-linux-x64.tar.xz`
    * `rm -f node`
    * `xz -cd node-v${NODEJS_VERSION}-linux-x64.tar.xz | tar xf -`
    * `ln -s node-v${NODEJS_VERSION}-linux-x64 node`
* PATH erweitern in ~/.bashrc:
    ```diff
    +PATH="${HOME}/Software/node/bin:${PATH}"
    +export PATH
    ```

Bauen
-----

Bei mir erfordert das Bauen
Tätigkeiten am lokalen Arbeitsplatz
und am Build-Container.

### Lokaler Arbeitsplatz

#### Ablauf  bei der ersten Version

* Basisversion wählen: v1.13.2
* Basisversion auschecken: `git checkout -b 1.13.2-uli v1.13.2`
* Zusammenführen mit "anderen" Versionen
    * `git log ed25519_sk`
    * `git cherry-pick 406fa489ba80f73e879f9336ee1689eb2eecccdb`
* Eventuell noch zusätzliche Änderungen vornehmen
* Version hochschieben nach Github: `git push -u origin 1.13.2-uli`
* Tag erzeugen: `git tag 1.13.2-uli-01` 
* Tag hochschieben: `git push --tags`

#### Ablauf bei Folgeversionen

Voraussetzung: Es gibt bereits einen Zweig (branch) mit allen
gewünschten Modifikationen. Wir wollen diese nur für einen neuen
Basiszweig übernehmen!

```
#
# - Kommandozeile öffnen
# - `bash` ausführen
# - Nachfolgende Zeilen reinkopieren
# - Auf Fehler kontrollieren
# - `exit` ausführen
#
set -e
OLD_BASE=v1.15.4
NEW_BASE=v1.15.5
OLD_ULI="$(echo "${OLD_BASE}"|cut -c2-)-uli"
NEW_ULI="$(echo "${NEW_BASE}"|cut -c2-)-uli"
git checkout "${OLD_ULI}"
OLD_TAG="$(git describe --tags "$(git rev-list --tags --max-count=1)")"
test "$(git describe "${OLD_TAG}")" != "$(git describe "${OLD_ULI}")" && {
  # Create a new tag for the old base
  OLD_TAG_COUNT="$(echo "${OLD_TAG}"|sed -e "s/^${OLD_ULI}-//")"
  INCREMENTED_COUNT="$(printf "%02d" "$(expr "${OLD_TAG_COUNT}" + 1)")"
  OLD_TAG2="$(echo "${OLD_TAG}"|sed -e "s/-${OLD_TAG_COUNT}$/-${INCREMENTED_COUNT}/")"
  git tag "${OLD_TAG2}"
  git push --tags
  OLD_TAG="${OLD_TAG2}"
}
git rebase "${NEW_BASE}"
git checkout -b "${NEW_ULI}"
git push -u origin "${NEW_ULI}"
NEW_TAG="$(echo "${OLD_TAG}"|sed -e "s/^${OLD_ULI}-/${NEW_ULI}-/")"
git tag "${NEW_TAG}"
git push --tags
set +e
```

### Build-Container

* Anmelden mit `ssh -A...`, damit wir eine Verbindung zu GITHUB bekommen
* Alle Dinge von GITHUB abholen: `git fetch --all -p` -> neues Tag wird angezeigt
* Tag auschecken: `git checkout 1.13.2-uli-01` -> Warnung bzgl. 'detached HEAD' ignorieren
* Versionstest: `git describe --tags --always` -> neues Tag wird angezeigt
* Bauen:
    ```
    make clean
    TAGS="bindata sqlite sqlite_unlock_notify" make build
    ```
* Erneuter Versionstest: `./gitea --version` -> "Gitea version 1.13.2+uli-01 built with..."
* Artefakt zum Hochladen erzeugen: `xz -c9 gitea >gitea-1.13.2-uli-01-linux-amd64.xz`
* Artefakt in Github ablegen und lokal löschen
