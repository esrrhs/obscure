#!/usr/bin/env python3
"""Batch generate bg/portrait. Default engine: SiliconFlow Kolors (same as TheMandate)."""
import os, sys, json, base64, time, urllib.request
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
CONTENT = ROOT / "content"

# Load local .env without exporting into shell history dumps
def _load_dotenv():
    env_path = ROOT / ".env"
    if not env_path.exists():
        return
    for line in env_path.read_text(encoding="utf-8").splitlines():
        line = line.strip()
        if not line or line.startswith("#") or "=" not in line:
            continue
        k, v = line.split("=", 1)
        os.environ.setdefault(k.strip(), v.strip())

_load_dotenv()

ENGINE = os.environ.get("IMAGE_ENGINE", "kolors").lower()  # kolors | freellmapi
KOLORS_KEY = os.environ.get("KOLORS_API_KEY", "")
KOLORS_URL = "https://api.siliconflow.cn/v1/images/generations"
KOLORS_MODEL = "Kwai-Kolors/Kolors"
KOLORS_SIZE = "2016x1120"

FREE_URL = os.environ.get("FREELLMAPI_URL", "http://127.0.0.1:13001/v1/images/generations")
FREE_KEY = os.environ.get("FREELLMAPI_API_KEY", "")
FREE_MODEL = os.environ.get("IMAGE_MODEL", "black-forest-labs/flux.1-dev")

BG_SIZE = KOLORS_SIZE if ENGINE == "kolors" else "1344x768"
PORTRAIT_SIZE = "768x1024"
SHARP_SUFFIX = ", sharp focus, highly detailed, crisp brushwork, clear edges, ultra detailed"


def crop_to_aspect(path: Path, aspect: float, keep_png: bool = False) -> Path | None:
    try:
        from PIL import Image
    except ImportError:
        return None
    im = Image.open(path).convert("RGB")
    w, h = im.size
    cur = w / h
    if abs(cur - aspect) >= 0.02:
        if cur > aspect:
            nw = int(h * aspect)
            left = (w - nw) // 2
            im = im.crop((left, 0, left + nw, h))
        else:
            nh = int(w / aspect)
            top = (h - nh) // 2
            im = im.crop((0, top, w, top + nh))
    if keep_png:
        out = path if path.suffix.lower() == ".png" else path.with_suffix(".png")
        im.save(out, "PNG", optimize=True)
    else:
        # Kolors returns PNG; save as high-quality JPEG to keep repo smaller
        out = path if path.suffix.lower() in {".jpg", ".jpeg"} else path.with_suffix(".jpg")
        im.save(out, quality=93, optimize=True)
    if out != path and path.exists():
        path.unlink()
    return out


def make_transparent_portrait(src: Path, dst: Path) -> Path:
    """Remove background → RGBA PNG, harden alpha, crop to content."""
    from PIL import Image
    from rembg import remove

    im = Image.open(src).convert("RGBA")
    cut = remove(im)
    r, g, b, a = cut.split()
    a = a.point(lambda x: 0 if x < 40 else (255 if x > 180 else x))
    out = Image.merge("RGBA", (r, g, b, a))
    bbox = out.getbbox()
    if bbox:
        out = out.crop(bbox)
    dst = dst.with_suffix(".png")
    for p in dst.parent.glob(dst.stem + ".*"):
        if p.suffix.lower() in {".png", ".jpg", ".jpeg", ".webp"} and p != dst:
            p.unlink(missing_ok=True)
    out.save(dst, "PNG", optimize=True)
    print(f"  [透明立绘] {dst} ({out.size[0]}x{out.size[1]})")
    return dst


def mean_luma(path: Path) -> float:
    from PIL import Image
    im = Image.open(path).convert("L")
    px = list(im.getdata())
    return sum(px) / max(len(px), 1)


def _write_image(raw: bytes, dst: Path, aspect: float | None, portrait: bool = False) -> Path:
    dst.parent.mkdir(parents=True, exist_ok=True)
    if raw[:8] == b"\x89PNG\r\n\x1a\n":
        tmp = dst.with_suffix(".png")
    elif raw[:3] == b"\xff\xd8\xff":
        tmp = dst.with_suffix(".jpg")
    else:
        tmp = dst
    for p in tmp.parent.glob(dst.stem + ".*"):
        if p.suffix.lower() in {".png", ".jpg", ".jpeg", ".webp"}:
            p.unlink()
    tmp.write_bytes(raw)
    if aspect:
        out = crop_to_aspect(tmp, aspect, keep_png=portrait) or tmp
    else:
        out = tmp
    if isinstance(out, Path):
        final = out
    else:
        final = tmp.with_suffix(".jpg") if tmp.with_suffix(".jpg").exists() else tmp
    if not final.exists():
        cands = list(dst.parent.glob(dst.stem + ".*"))
        if not cands:
            raise RuntimeError("image missing after write")
        final = cands[0]
    if portrait:
        final = make_transparent_portrait(final, dst.with_suffix(".png"))
    m = mean_luma(final)
    if m < 8:
        final.unlink(missing_ok=True)
        raise RuntimeError(f"image too dark mean={m:.1f}")
    print(f"  [OK] {final} ({final.stat().st_size} bytes, mean={m:.1f})")
    return final


def gen_kolors(prompt: str, dst: Path, retries: int = 5) -> None:
    if not KOLORS_KEY:
        raise RuntimeError("未设置 KOLORS_API_KEY（可写在项目根 .env，勿提交）")
    is_portrait = dst.name.startswith("portrait")
    size = PORTRAIT_SIZE if is_portrait else KOLORS_SIZE
    body = json.dumps({
        "model": KOLORS_MODEL,
        "prompt": prompt + SHARP_SUFFIX,
        "image_size": size,
        "num_inference_steps": 30,
        "guidance_scale": 7.5,
    }).encode()
    wait = 2
    aspect = 3 / 4 if is_portrait else 16 / 9
    for attempt in range(1, retries + 1):
        try:
            req = urllib.request.Request(
                KOLORS_URL, data=body, method="POST",
                headers={"Authorization": f"Bearer {KOLORS_KEY}", "Content-Type": "application/json"},
            )
            with urllib.request.urlopen(req, timeout=180) as resp:
                data = json.load(resp)
            imgs = data.get("images") or data.get("data") or []
            if not imgs or not imgs[0].get("url"):
                raise RuntimeError(f"no url: {str(data)[:200]}")
            with urllib.request.urlopen(imgs[0]["url"], timeout=120) as r:
                raw = r.read()
            _write_image(raw, dst, aspect, portrait=is_portrait)
            return
        except Exception as e:
            print(f"  [Kolors 失败 {attempt}/{retries}] {e}")
            if attempt == retries:
                raise
            time.sleep(wait)
            wait = min(wait * 2, 60)


def gen_freellmapi(prompt: str, size: str, dst: Path, retries: int = 5) -> None:
    if not FREE_KEY:
        raise RuntimeError("未设置 FREELLMAPI_API_KEY")
    is_portrait = dst.name.startswith("portrait")
    body = json.dumps({
        "model": FREE_MODEL,
        "prompt": prompt + SHARP_SUFFIX,
        "size": size,
        "n": 1,
        "response_format": "b64_json",
    }).encode()
    wait = 2
    aspect = 3 / 4 if is_portrait else (16 / 9 if dst.name.startswith("bg") or "cover" in str(dst) else None)
    for attempt in range(1, retries + 1):
        try:
            req = urllib.request.Request(
                FREE_URL, data=body, method="POST",
                headers={"Authorization": f"Bearer {FREE_KEY}", "Content-Type": "application/json"},
            )
            with urllib.request.urlopen(req, timeout=180) as resp:
                data = json.load(resp)
            item = (data.get("data") or [None])[0]
            if not item:
                raise RuntimeError("empty data")
            if "b64_json" in item:
                raw = base64.b64decode(item["b64_json"])
            elif "url" in item:
                with urllib.request.urlopen(item["url"], timeout=120) as r:
                    raw = r.read()
            else:
                raise RuntimeError("no image payload")
            _write_image(raw, dst, aspect, portrait=is_portrait)
            return
        except Exception as e:
            print(f"  [freellmapi 失败 {attempt}/{retries}] {e}")
            if attempt == retries:
                raise
            time.sleep(wait)
            wait = min(wait * 2, 60)


def gen(prompt: str, size: str, dst: Path) -> None:
    if ENGINE == "kolors":
        gen_kolors(prompt, dst)
    else:
        gen_freellmapi(prompt, size, dst)


def main():
    mode = sys.argv[1] if len(sys.argv) > 1 else "bg"
    max_n = int(sys.argv[2]) if len(sys.argv) > 2 else 999
    print(f"[engine={ENGINE}]")
    if ENGINE == "kolors" and not KOLORS_KEY:
        sys.exit("需要 KOLORS_API_KEY（写入 .env，勿提交到 git）")

    done = 0
    if mode in ("all", "bg"):
        scenes = sorted((CONTENT / "scenes").iterdir(), key=lambda p: int(p.name) if p.name.isdigit() else 999)
        for sc in scenes:
            if not sc.is_dir() or not sc.name.isdigit():
                continue
            img_txt = sc / "img.txt"
            if not img_txt.exists():
                continue
            if any(sc.glob("bg.*")):
                print(f"[跳过幕 {sc.name}] 已有背景")
                continue
            prompt = img_txt.read_text(encoding="utf-8").strip()
            print(f"[幕 {sc.name}] 生成背景…")
            gen(prompt, BG_SIZE, sc / "bg.jpg")
            done += 1
            if done >= max_n:
                print(f"达到上限 {max_n}")
                return
            time.sleep(1.0)

    if mode in ("all", "char"):
        for ch in sorted((CONTENT / "characters").iterdir()):
            if not ch.is_dir():
                continue
            img_txt = ch / "img.txt"
            if not img_txt.exists():
                continue
            if any(ch.glob("portrait.*")):
                print(f"[跳过立绘 {ch.name}] 已有")
                continue
            prompt = img_txt.read_text(encoding="utf-8").strip()
            print(f"[立绘 {ch.name}] 生成正面透明 PNG…")
            gen(prompt, PORTRAIT_SIZE, ch / "portrait.png")
            done += 1
            if done >= max_n:
                print(f"达到上限 {max_n}")
                return
            time.sleep(1.0)

    if mode in ("all", "cover", "bg"):
        cover = ROOT / "cover.jpg"
        if mode == "bg" and cover.exists() and cover.stat().st_size > 1000 and done:
            pass  # only regenerate cover when missing or in cover/all
        need_cover = mode in ("all", "cover") or not cover.exists() or cover.stat().st_size < 1000
        # when regenerating all bgs, also redo cover
        if mode == "bg" and done > 0:
            # if we cleared covers separately, handle below
            pass
        if mode in ("all", "cover") or (mode == "bg" and (not cover.exists() or cover.stat().st_size < 1000)):
            prompt_file = ROOT / "cover_prompt.txt"
            prompt = prompt_file.read_text(encoding="utf-8").strip() if prompt_file.exists() else \
                (CONTENT / "scenes" / "1" / "img.txt").read_text(encoding="utf-8").strip()
            print("[封面] 生成…")
            gen(prompt, BG_SIZE, cover)
            done += 1

    print(f"完成，本次新生成约 {done} 张")


if __name__ == "__main__":
    main()
