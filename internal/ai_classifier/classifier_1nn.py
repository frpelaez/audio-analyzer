import glob
import json
import os



class AudioFingerprintClassifier:
    def __init__(self):
        self.database = {}  # Estructura: {'nombre_cancion': set(frecuencias)}

    def load_fingerprint(self, filepath):
        """
        Lee el JSON y extrae un CONJUNTO (set) de frecuencias únicas.
        Ignoramos el tiempo (offset) como se solicitó.
        """
        try:
            with open(filepath, "r", encoding="utf-8") as f:
                data = json.load(f)

            # Asumimos que la estructura del JSON es la que generaste en Go:
            # {"points": [{"freq": 1024, "time": ...}, ...]}
            # Extraemos solo las frecuencias y las convertimos a enteros para comparacion exacta
            points = data.get("points", [])

            # Usamos un set para eliminar duplicados y hacer búsquedas O(1)
            # Esto crea la "Huella Espectral Global" de la canción
            frequencies = set(
                int(p["FreqHz"]) if "FreqHz" in p else int(p["f"]) for p in points
            )

            return frequencies, len(points)
        except Exception as e:
            print(f"⚠️ Error leyendo {filepath}: {e}")
            return set(), 0

    def fit(self, db_folder):
        """
        'Entrena' el modelo cargando la base de datos en memoria.
        """
        print(f"📚 Cargando base de datos desde '{db_folder}'...")
        json_files = glob.glob(os.path.join(db_folder, "*.json"))

        for path in json_files:
            filename = os.path.basename(path)
            # El nombre de la clase es el nombre del archivo (sin extensión)
            song_name = os.path.splitext(filename)[0]

            features, _ = self.load_fingerprint(path)
            if features:
                self.database[song_name] = features

        print(f"✅ Modelo cargado con {len(self.database)} canciones de referencia.\n")

    def predict(self, fragment_path):
        """
        Clasifica un fragmento comparándolo contra toda la base de datos.
        """
        fragment_features, total_points = self.load_fingerprint(fragment_path)

        if not fragment_features:
            return None, 0.0

        best_song = None
        best_score = -1.0

        # --- Lógica de Clasificación (1-NN) ---
        for song_name, song_features in self.database.items():

            # Calculamos la INTERSECCIÓN (puntos en común)
            common_points = fragment_features.intersection(song_features)

            # Métrica: CONTAINMENT (Contención)
            # ¿Qué porcentaje de las frecuencias del fragmento existen en la canción original?
            # Fórmula: |A ∩ B| / |A|  (Donde A es el fragmento)
            if len(fragment_features) > 0:
                score = len(common_points) / len(fragment_features)
            else:
                score = 0

            if score > best_score:
                best_score = score
                best_song = song_name

        return best_song, best_score

    def evaluate_batch(self, test_folder):
        print(f"🕵️  Iniciando clasificación de fragmentos en '{test_folder}'...")
        test_files = glob.glob(os.path.join(test_folder, "*.json"))

        results = []

        print(f"{'FRAGMENTO':<40} | {'PREDICCIÓN':<40} | {'CONFIANZA'}")
        print("-" * 100)

        for path in test_files:
            fragment_name = os.path.basename(path)
            prediction, score = self.predict(path)

            print(f"{fragment_name:<40} | {prediction:<40} | {score:.2%}")
            results.append((fragment_name, prediction, score))

        return results


# --- EJECUCIÓN DEL PROGRAMA ---

if __name__ == "__main__":
    # Configura aquí tus rutas
    DB_FOLDER = "db"  # Carpeta con los JSON completos
    TEST_FOLDER = "fdb"  # Carpeta con los JSON de los fragmentos

    # 1. Instanciar Modelo
    clf = AudioFingerprintClassifier()

    # 2. Cargar Datos (Entrenamiento)
    if os.path.exists(DB_FOLDER):
        clf.fit(DB_FOLDER)
    else:
        print(f"❌ La carpeta {DB_FOLDER} no existe.")
        exit()

    # 3. Predecir (Test)
    if os.path.exists(TEST_FOLDER):
        clf.evaluate_batch(TEST_FOLDER)
    else:
        print(
            f"⚠️ La carpeta {TEST_FOLDER} no existe. Crea algunos JSON de prueba primero."
        )
