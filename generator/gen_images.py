#!/usr/bin/env python3
"""Batch generate bg/portrait via freellmapi OpenAI-compatible images API."""
import os, sys, json, base64, time, urllib.request, urllib.error
from pathlib import Path

API = os.environ.get("FREELLMAPI_URL", "http://127.0.0.1:13001/v1/images/generations")
KEY = os.environ.get("FREELLMAPI_API_KEY", "")
MODEL = os.environ.get("IMAGE_MODEL", "black-forest-labs/flux.1-dev")
ROOT = Path(__file__).resolve().parents[1]
CONTENT = ROOT / "content"

BG_SIZE = "1344x768"
PORTRAIT_SIZE = "768x1024"


def crop_to_aspect(path: Path, aspect: float) -> None:
    """Center-crop image to target aspect ratio using stdlib only (JPEG/PNG via raw is hard);
    use Pillow if available, else skip."""
    try:
        from PIL import Image
    except ImportError:
        return
    im = Image.open(path)
    w, h = im.size
    cur = w / h
    if abs(cur - aspect) < 0.02:
        return
    if cur > aspect:
        # too wide
        nw = int(h * aspect)
        left = (w - nw) // 2
        im = im.crop((left, 0, left + nw, h))
    else:
        nh = int(w / aspect)
        top = (h - nh) // 2
        im = im.crop((0, top, w, top + nh))
    im.save(path, quality=92)

def gen(prompt: str, size: str, dst: Path, retries: int = 5) -> None:
    if dst.exists() and dst.stat().st_size > 1000:
        print(f"  [跳过] {dst}")
        return
    body = json.dumps({
        "model": MODEL,
        "prompt": prompt,
        "size": size,
        "n": 1,
        "response_format": "b64_json",
    }).encode()
    wait = 2
    for attempt in range(1, retries + 1):
        req = urllib.request.Request(
            API, data=body, method="POST",
            headers={
                "Authorization": f"Bearer {KEY}",
                "Content-Type": "application/json",
            },
        )
        try:
            with urllib.request.urlopen(req, timeout=180) as resp:
                data = json.load(resp)
            items = data.get("data") or []
            if not items:
                raise RuntimeError(f"empty data: {data}")
            item = items[0]
            if "b64_json" in item:
                raw = base64.b64decode(item["b64_json"])
            elif "url" in item:
                with urllib.request.urlopen(item["url"], timeout=120) as r:
                    raw = r.read()
            else:
                raise RuntimeError(f"no image payload: {item.keys()}")
            dst.parent.mkdir(parents=True, exist_ok=True)
            # freellmapi/nvidia returns jpeg bytes even if we ask png
            if raw[:3] == b"\xff\xd8\xff":
                out = dst.with_suffix(".jpg")
            elif raw[:8] == b"\x89PNG\r\n\x1a\n":
                out = dst.with_suffix(".png")
            else:
                out = dst
            # remove sibling ext leftovers
            for p in out.parent.glob(dst.stem + ".*"):
                if p.suffix in {".png", ".jpg", ".jpeg", ".webp"} and p != out:
                    p.unlink()
            out.write_bytes(raw)
            # Prefer 16:9 for backgrounds / cover
            if out.name.startswith("bg") or out.name.startswith("cover") or "cover" in str(out):
                try:
                    crop_to_aspect(out, 16/9)
                except Exception as ce:
                    print(f"  [裁切警告] {ce}")
            elif out.name.startswith("portrait"):
                try:
                    crop_to_aspect(out, 3/4)
                except Exception as ce:
                    print(f"  [裁切警告] {ce}")
            # reject near-black / empty images
            try:
                from PIL import Image as _Image
                _im = _Image.open(out).convert("L")
                _px = list(_im.getdata())
                _mean = sum(_px) / max(len(_px), 1)
                if _mean < 8:
                    out.unlink(missing_ok=True)
                    raise RuntimeError(f"image too dark mean={_mean:.1f}, treat as failed")
            except RuntimeError:
                raise
            except Exception:
                pass
            print(f"  [OK] {out} ({out.stat().st_size} bytes)")
            return
        except Exception as e:
            print(f"  [失败 {attempt}/{retries}] {e}")
            if attempt == retries:
                print("  [改用 pollinations 兜底]")
                gen_pollinations(prompt, size, dst)
                return
            time.sleep(wait)
            wait = min(wait * 2, 60)


def gen_pollinations(prompt: str, size: str, dst: Path) -> None:
    import urllib.parse
    w, h = 1344, 768
    if "x" in size:
        try:
            w, h = map(int, size.lower().split("x"))
        except Exception:
            pass
    url = (
        "https://image.pollinations.ai/prompt/"
        + urllib.parse.quote(prompt[:450])
        + f"?width={w}&height={h}&nologo=true&model=flux&enhance=true"
    )
    with urllib.request.urlopen(url, timeout=180) as resp:
        raw = resp.read()
    if len(raw) < 2000:
        raise RuntimeError("pollinations returned tiny payload")
    dst.parent.mkdir(parents=True, exist_ok=True)
    out = dst.with_suffix(".jpg")
    for p in out.parent.glob(dst.stem + ".*"):
        if p.suffix in {".png", ".jpg", ".jpeg", ".webp"} and p != out:
            p.unlink()
    out.write_bytes(raw)
    if out.name.startswith("bg") or "cover" in str(out):
        try:
            crop_to_aspect(out, 16 / 9)
        except Exception:
            pass
    elif out.name.startswith("portrait"):
        try:
            crop_to_aspect(out, 3 / 4)
        except Exception:
            pass
    from PIL import Image as _Image
    _im = _Image.open(out).convert("L")
    _mean = sum(_im.getdata()) / max(_im.size[0] * _im.size[1], 1)
    if _mean < 8:
        out.unlink(missing_ok=True)
        raise RuntimeError(f"pollinations too dark mean={_mean:.1f}")
    print(f"  [OK/poll] {out} ({out.stat().st_size} bytes)")

def main():
    if not KEY:
        sys.exit("需要 FREELLMAPI_API_KEY")
    mode = sys.argv[1] if len(sys.argv) > 1 else "all"
    max_n = int(sys.argv[2]) if len(sys.argv) > 2 else 999

    done = 0
    if mode in ("all", "bg"):
        scenes = sorted((CONTENT / "scenes").iterdir(), key=lambda p: int(p.name) if p.name.isdigit() else 999)
        for sc in scenes:
            if not sc.is_dir() or not sc.name.isdigit():
                continue
            img_txt = sc / "img.txt"
            if not img_txt.exists():
                continue
            # skip if any bg.*
            if any(sc.glob("bg.*")):
                print(f"[跳过幕 {sc.name}] 已有背景")
                continue
            prompt = img_txt.read_text(encoding="utf-8").strip()
            print(f"[幕 {sc.name}] 生成背景…")
            gen(prompt, BG_SIZE, sc / "bg.png")
            done += 1
            if done >= max_n:
                print(f"达到上限 {max_n}")
                return
            time.sleep(1.5)

    if mode in ("all", "char"):
        chars = sorted((CONTENT / "characters").iterdir())
        for ch in chars:
            if not ch.is_dir():
                continue
            img_txt = ch / "img.txt"
            if not img_txt.exists():
                continue
            if any(ch.glob("portrait.*")):
                print(f"[跳过立绘 {ch.name}] 已有")
                continue
            prompt = img_txt.read_text(encoding="utf-8").strip()
            print(f"[立绘 {ch.name}] 生成…")
            gen(prompt, PORTRAIT_SIZE, ch / "portrait.png")
            done += 1
            if done >= max_n:
                print(f"达到上限 {max_n}")
                return
            time.sleep(1.5)

    # cover
    if mode in ("all", "cover"):
        cover = ROOT / "cover.jpg"
        if not cover.exists() or cover.stat().st_size < 1000:
            prompt = (CONTENT / "scenes" / "27" / "img.txt").read_text(encoding="utf-8").strip()
            # prefer a cover-specific prompt
            cover_prompt_file = ROOT / "cover_prompt.txt"
            if cover_prompt_file.exists():
                prompt = cover_prompt_file.read_text(encoding="utf-8").strip()
            print("[封面] 生成…")
            gen(prompt, BG_SIZE, cover)
        else:
            print("[跳过封面]")

    print(f"完成，本次新生成约 {done} 张")

if __name__ == "__main__":
    main()
