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

```
cd ~/git/forked/gitea

```
