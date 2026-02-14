import glob
import json
import os
from collections import Counter, defaultdict



class TemporalFingerprintClassifier:
    def __init__(self):
        # Estructura:
        # {
        #   'nombre_cancion': {
        #       frecuencia_1: [t1, t3, t8],
        #       frecuencia_2: [t2]
        #   }
        # }
        self.database = {}

    def load_fingerprint_data(self, filepath):
        """
        Lee el JSON y extrae una lista de tuplas (frecuencia, tiempo).
        """
        try:
            with open(filepath, "r", encoding="utf-8") as f:
                data = json.load(f)

            points = data.get("points", [])
            # Extraemos (freq, time)
            # Normalizamos frecuencia a entero, mantenemos tiempo como float
            extracted_points = []
            for p in points:
                freq = int(p["FreqHz"]) if "FreqHz" in p else int(p["f"])
                time = float(p["TimeSec"]) if "TimeSec" in p else float(p["t"])
                extracted_points.append((freq, time))

            return extracted_points
        except Exception as e:
            print(f"⚠️ Error leyendo {filepath}: {e}")
            return []

    def fit(self, db_folder):
        """
        Carga la DB indexando por frecuencias para búsqueda rápida.
        """
        print(
            f"📚 Cargando base de datos con estructura temporal desde '{db_folder}'..."
        )
        json_files = glob.glob(os.path.join(db_folder, "*.json"))

        for path in json_files:
            filename = os.path.basename(path)
            song_name = os.path.splitext(filename)[0]

            points = self.load_fingerprint_data(path)

            # Indexamos: Frecuencia -> Lista de Tiempos donde aparece
            song_index = defaultdict(list)
            for freq, time in points:
                song_index[freq].append(time)

            self.database[song_name] = song_index

        print(f"✅ Modelo cargado con {len(self.database)} canciones.\n")

    def predict(self, fragment_path):
        """
        Clasifica verificando la COHERENCIA TEMPORAL.
        """
        fragment_points = self.load_fingerprint_data(fragment_path)
        if not fragment_points:
            return None, 0.0

        best_song = None
        best_coherence = -1

        # Iteramos sobre cada canción candidata en la base de datos
        for song_name, song_index in self.database.items():

            time_differences = []

            # 1. Alineación
            # Para cada punto del fragmento, buscamos si existe esa frecuencia en la canción
            for f_freq, f_time in fragment_points:
                if f_freq in song_index:
                    # Si la frecuencia existe, calculamos la diferencia de tiempo (offset)
                    # con TODAS las veces que esa frecuencia aparece en la canción original.
                    for song_time in song_index[f_freq]:
                        delta = song_time - f_time
                        # Redondeamos para tolerar pequeñas desviaciones (jitter)
                        # round(x, 1) agrupa en bins de 100ms
                        time_differences.append(round(delta, 1))

            if not time_differences:
                continue

            # 2. Histograma de Diferencias (Voting)
            # Buscamos cuál es la diferencia de tiempo más común (el Offset implícito)
            # Si es un match real, habrá un pico enorme en un solo valor.
            # Si es ruido, los valores estarán dispersos.
            offset_counts = Counter(time_differences)

            # Obtenemos el número de votos del offset ganador
            if offset_counts:
                max_votes = offset_counts.most_common(1)[0][1]
            else:
                max_votes = 0

            # Puntuación de Coherencia
            if max_votes > best_coherence:
                best_coherence = max_votes
                best_song = song_name

        # Normalizamos el score (0.0 a 1.0) basado en el tamaño del fragmento
        final_score = 0.0
        if len(fragment_points) > 0:
            final_score = best_coherence / len(fragment_points)

        return best_song, final_score

    def evaluate_batch(self, test_folder):
        print(f"🕵️  Analizando coherencia temporal en '{test_folder}'...")
        test_files = glob.glob(os.path.join(test_folder, "*.json"))

        print(f"{'FRAGMENTO':<30} | {'PREDICCIÓN':<30} | {'COHERENCIA'}")
        print("-" * 80)

        for path in test_files:
            fragment_name = os.path.basename(path)
            prediction, score = self.predict(path)

            tag = ""
            if score > 0.05:
                tag = "✅"  # Umbral visual

            print(f"{fragment_name:<30} | {prediction:<30} | {score:.2%} {tag}")


# --- MAIN ---
if __name__ == "__main__":
    DB_FOLDER = "db"
    TEST_FOLDER = "fdb"  # Asegúrate de que esta carpeta tenga los JSONs de los fragmentos

    clf = TemporalFingerprintClassifier()
    if os.path.exists(DB_FOLDER):
        clf.fit(DB_FOLDER)
        if os.path.exists(TEST_FOLDER):
            clf.evaluate_batch(TEST_FOLDER)
    else:
        print("❌ Crea las carpetas primero.")
