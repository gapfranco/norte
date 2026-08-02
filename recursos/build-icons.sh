#!/bin/sh
set -eu

ROOT="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
SRC="$ROOT/recursos/logos/norte-icon.svg"
OUT="$ROOT/ui/static/icons"

mkdir -p "$OUT"
cp "$SRC" "$OUT/norte-icon.svg"

have_all_icons() {
  [ -f "$OUT/norte-icon-32.png" ] &&
    [ -f "$OUT/norte-icon-180.png" ] &&
    [ -f "$OUT/norte-icon-192.png" ] &&
    [ -f "$OUT/norte-icon-512.png" ] &&
    [ -f "$OUT/norte.ico" ]
}

INKSCAPE_BIN=""
if command -v inkscape >/dev/null 2>&1; then
  INKSCAPE_BIN=inkscape
fi

MAGICK_BIN=""
if command -v magick >/dev/null 2>&1; then
  MAGICK_BIN=magick
elif command -v convert >/dev/null 2>&1; then
  MAGICK_BIN=convert
fi

export_png() {
  size="$1"
  dest="$OUT/norte-icon-${size}.png"
  # Inkscape renders <text> reliably; prefer it over ImageMagick for this SVG.
  if [ -n "$INKSCAPE_BIN" ]; then
    inkscape "$SRC" --export-type=png --export-filename="$dest" -w "$size" -h "$size"
    return
  fi
  if [ "$MAGICK_BIN" = "magick" ]; then
    magick -background none -density 300 "$SRC" -resize "${size}x${size}" "$dest"
    return
  fi
  if [ "$MAGICK_BIN" = "convert" ]; then
    convert -background none -density 300 "$SRC" -resize "${size}x${size}" "$dest"
    return
  fi
  return 1
}

if [ -z "$INKSCAPE_BIN" ] && [ -z "$MAGICK_BIN" ]; then
  if have_all_icons; then
    echo "Inkscape/ImageMagick não encontrados; reusando ícones já em ui/static/icons/"
    exit 0
  fi
  echo "erro: instale Inkscape (recomendado) ou ImageMagick, ou commite os PNGs/ICO." >&2
  exit 1
fi

for size in 32 180 192 512; do
  export_png "$size"
  echo "gerado: norte-icon-${size}.png"
done

if [ -z "$MAGICK_BIN" ]; then
  if [ -f "$OUT/norte.ico" ]; then
    echo "ImageMagick ausente; reusando norte.ico existente"
    exit 0
  fi
  echo "erro: ImageMagick necessário para gerar norte.ico" >&2
  exit 1
fi

if [ "$MAGICK_BIN" = "magick" ]; then
  magick "$OUT/norte-icon-32.png" "$OUT/norte.ico"
else
  convert "$OUT/norte-icon-32.png" "$OUT/norte.ico"
fi
echo "gerado: norte.ico"
