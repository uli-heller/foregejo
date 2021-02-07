Uli's Bau-Anleitung
===================

Lokaler Arbeitsplatz
--------------------

* Basisversion wählen: v1.13.2
* Basisversion auschecken: `git checkout -b 1.13.2-uli v1.13.2`
* Zusammenführen mit "anderen" Versionen
    * `git log ed25519_sk`
    * `git cherry-pick 406fa489ba80f73e879f9336ee1689eb2eecccdb`
* Eventuell noch zusätzliche Änderungen vornehmen
* Version hochschieben nach Github: `git push -u origin 1.13.2-uli`
* Tag erzeugen: `git tag 1.13.2-uli-01` 
* Tag hochschieben: `git push --tags`

Build-Container
---------------

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
