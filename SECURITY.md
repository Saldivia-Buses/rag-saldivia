# Política de seguridad

## Reportar una vulnerabilidad

**No abras un issue público** para informes de seguridad.

Enviá un correo al mantenedor del repositorio con:

- Descripción del problema y posible impacto
- Versión o commit afectado (branch o tag)
- Pasos mínimos para reproducir
- Si tenés una idea de mitigación, opcional

Trataremos de responder en **best effort** (sin SLA formal en un proyecto de equipo reducido).

## Secretos y entorno

Variables sensibles:

- `JWT_SECRET` — nunca commitear; rotar si se filtra
- `REDIS_URL` — puede incluir credenciales; no exponer en logs públicos
- `SYSTEM_API_KEY` — autenticación máquina-a-máquina; rotar si se filtra

## Qué no commitear

- `.env`, `.env.local`, `*.env` con secretos reales
- Archivos `*.key`, certificados o PEM privados
- Credenciales hardcodeadas en código fuente

Usá `.env.example` solo con placeholders y documentación.
