import os

import yt_dlp

input = "songs.txt"
output = "audios_test"


def download_songs_from_file(file: str):
    ydl_opts = {
        "format": "bestaudio/best",
        "postprocessors": [
            {
                "key": "FFmpegExtractAudio",
                "preferredcodec": "wav",
                "preferredquality": "192",
            }
        ],
        "outtmpl": f"{output}/%(title)s.%(ext)s",
        "default_search": "ytsearch1",
        "quiet": False,
    }

    if not os.path.exists(file):
        print(f"Error: file not found {file}")
        return

    if not os.path.exists(output):
        os.makedirs(output)

    with yt_dlp.YoutubeDL(ydl_opts) as ydl:
        with open(file, "r", encoding="utf-8") as f:
            for line in f:
                search = line.strip()
                if search:
                    print(f"\nDownloading: {search}")
                    try:
                        ydl.download([search])
                    except Exception as e:
                        print(f"Error with {search}: {e}")


if __name__ == "__main__":
    download_songs_from_file(input)
