from pathlib import Path
from fontTools.ttLib import TTFont
from fontTools.pens.recordingPen import DecomposingRecordingPen
from fontTools.pens.transformPen import TransformPen
from fontTools.pens.ttGlyphPen import TTGlyphPen

root = Path(__file__).resolve().parents[1]
font = TTFont(root / 'assets' / 'IBMPlexSansArabic-Regular.ttf')
glyphs = font.getGlyphSet()
names = {name for code, name in font.getBestCmap().items() if 0x600 <= code <= 0x8ff or 0xfb50 <= code <= 0xfdff or 0xfe70 <= code <= 0xfeff}
changed = {}
scale = 1.30
for name in names:
    recording = DecomposingRecordingPen(glyphs)
    glyphs[name].draw(recording)
    pen = TTGlyphPen(None)
    recording.replay(TransformPen(pen, (scale, 0, 0, scale, 0, 0)))
    changed[name] = pen.glyph()
for name, glyph in changed.items():
    font['glyf'][name] = glyph
    advance, bearing = font['hmtx'][name]
    font['hmtx'][name] = (round(advance * scale), round(bearing * scale))
for name in font['name'].names:
    if name.nameID in (1, 3, 4, 6, 16, 18, 21, 25):
        value = 'VencordARUI-Regular' if name.nameID in (3, 6, 25) else 'Vencord AR UI Regular'
        name.string = value.encode(name.getEncoding())
font.save(root / 'assets' / 'VencordARUI-Regular.ttf')

