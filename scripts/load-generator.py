import urllib.request
import urllib.error
import time
import sys

TARGET_URL = "http://localhost/projeto-korp"
ERROR_URL = "http://localhost/rota-inexistente"

print("==> Iniciando gerador de carga simulada para validação do Grafana...")

# 1. Dispara 30 requisições de sucesso (HTTP 200)
print("[*] Enviando tráfego de sucesso (GET /projeto-korp)...")
for i in range(30):
    try:
        req = urllib.request.Request(TARGET_URL, method="GET")
        with urllib.request.urlopen(req) as response:
            pass
    except Exception as e:
        print(f"Erro na requisição {i}: {e}")
    time.sleep(0.1)

# 2. Dispara 5 requisições de erro 404 para acionar a métrica de erro no Prometheus/Grafana
print("[*] Simulando tráfego de erro 404 (GET /rota-inexistente)...")
for i in range(5):
    try:
        req = urllib.request.Request(ERROR_URL, method="GET")
        urllib.request.urlopen(req)
    except urllib.error.HTTPError as e:
        pass # Erro esperado 404
    except Exception as e:
        pass
    time.sleep(0.2)

print("==> Carga simulada concluída com sucesso! Verifique os gráficos no Grafana.")