---
# beans-376r
title: TUI settings overlay (preview position, preview ratio, tree mode)
status: todo
type: feature
priority: normal
created_at: 2026-10-05T15:31:09Z
updated_at: 2026-10-06T14:14:30Z
---

Ein Overlay, das sich mit `,` in der Liste und in der Detailansicht öffnet (gleiche Bedingung wie `?`: nur wenn kein anderes Overlay aktiv ist), stellt drei TUI-Einstellungen ein. Jede Änderung gilt sofort und wird nicht gespeichert: Sie gilt nur für die laufende Sitzung, `.beans.yml` bleibt unverändert.

```
preview at     ‹ auto | right | below ›
preview ratio  ‹ 50% ›        (nur bei Position below wirksam)
tree mode      ‹ all | active | unblocked | ready ›
```

**Tasten im Overlay:**
- `↑`/`↓` oder `k`/`j` wählt die Zeile.
- `←`/`→` oder `h`/`l` ändert den Wert. Die Ratio ändert sich in 10er-Schritten zwischen 10 und 90.
- `esc` oder `,` schließt das Overlay, `q` beendet die TUI.

Die bestehenden Tasten `v` (Preview-Position) und `h` (Tree-Modus) bleiben und ändern dieselben Werte. Hinter dem Overlay ist das echte Layout zu sehen, gedimmt, damit man die Ratio live einstellen kann.

## Regeln

**Ratio:**
- `p` ist der Anteil der Preview an der Terminalhöhe ohne Footer.
- Die Liste bekommt `round((height−1)·(100−p)/100)` Zeilen.
- Liste und Preview bekommen jeweils mindestens 5 Zeilen (`MinStackedListHeight`).
- `p` gilt nur bei Position `below`. Fällt `auto` oder `right` auf das gestapelte Layout zurück, gilt weiter `stacked_list_height` aus `.beans.yml`.

**Startwert und Resize:**
- `p` startet bei 50 (Liste und Preview 1:1).
- `p` bleibt bei jedem Resize als Faktor erhalten; die Höhen von Liste und Preview werden aus `p` und der neuen Terminalhöhe neu berechnet.

**Layoutwahl:**

| Position | Verhalten |
|---|---|
| `below` | Liste und Preview werden immer gestapelt, solange beide mindestens 5 Zeilen haben. Sonst wird nur die Liste angezeigt. Ein Fallback auf `right` gibt es nicht. |
| `right` | Zweispaltig ab 120 Spalten, sonst Fallback wie bisher. |
| `auto` | Wie bisher, mit der Schwelle `stacked_list_height + 20`; `p` wirkt nicht. |

## Todo

- [ ] Settings im `App` bündeln: Struct `tuiSettings{previewPosition, previewPercent}`. Die Methode `stackedListPaneHeight()` liefert bei `below` die aus `p` berechnete Höhe, sonst `stackedListHeight`. Sie ersetzt das Feld in `previewSize`, `listPaneSize`, `renderStackedView` und im Maus-Hit-Test.
- [ ] Layoutwahl in `isTwoColumnMode`/`isStackedMode` nach den Regeln umbauen. Bestehende Tests anpassen: Bei `below` hat die Liste bei 40 Zeilen `round(39·0,5) = 20` statt 15 Zeilen, und der Fallback von `below` auf `right` entfällt.
- [ ] Table-driven Tests: (Position, Breite, Höhe, p) → (Layout, Listenhöhe, Preview-Höhe), dazu ein Resize, bei dem `p` gleich bleibt und die Höhen neu berechnet werden. Jeder Fall prüft, dass `View()` genau `height` Zeilen hat.
- [ ] `renderListLayout()` aus `View()` auslagern und in `getBackgroundView()` nutzen, damit Modals das echte Layout gedimmt zeigen.
- [ ] `internal/tui/settingsoverlay.go` nach dem Muster von `help.go` anlegen: `viewSettingsOverlay`, `,` als Öffner, `settingsChangedMsg`/`closeSettingsMsg`, Neuladen bei einem Wechsel des Tree-Modus, `fitListToPane` bei einer Layoutänderung.
- [ ] Test für das Overlay: `,`, `↓`, `→`, `esc` senden, danach Settings, `list.viewMode` und `View()` prüfen.
- [ ] Eintrag `,  Settings` im Help-Overlay (`help.go`).
- [ ] `mise test`, danach manuell mit `mise beans tui` in einem kleinen und einem großen Terminal prüfen.
