# Flujo: Proyecto Simple

Flujo simplificado para proyectos pequeños o en etapas tempranas.

## Cuándo usar este flujo

Aplica cuando el proyecto cumple AL MENOS UNO de estos criterios:
- Menos de 50 archivos de código fuente
- Desarrollador único (no team)
- Etapa de prototipo o MVP
- Sin CI/CD configurado
- Sin capa de datos persistente
- Librería o CLI simple

Si el proyecto cumple 3 o más → flujo simple obligatorio.
Si cumple 1-2 → se puede elegir simple o completo según preferencia.

## Flujo simplificado

```
/sequoia init
  │
  └─► Detecta: proyecto simple
        ├─ Agentes reducidos: P1, P3, P4
        ├─ Opcionales: P2 (si hay bundle/API), P6 (si hay CI/deploy)
        ├─ Skip: P5 (no hay UI), P6 (sin infra)
        └─ Mode: quick por defecto

/sequoia audit --mode=quick
  │
  ├─ 1. Batch único (todos en paralelo)
  │     ├─ P1 Security  → hallazgos críticos y riesgos
  │     ├─ P3 Architecture → hallazgos críticos y riesgos
  │     ├─ P4 Quality → hallazgos críticos y riesgos (incluye deps)
  │     └─ [P6 Operations → solo si hay CI/deploy]
  │
  ├─ 2. Meta-agentes (simplificados)
  │     ├─ M1 Correlator → solo correlación directa, sin mapa de causas raíz
  │     └─ M2 Reporter → scores + output abreviado
  │
  └─ 3. Output único
        └─ sequoia-simple.md (todo en un archivo)
```

## Agentes reducidos: qué se salta

| Agente | Incluir | Motivo de exclusión |
|--------|---------|---------------------|
| P1 Security | ✅ | Siempre relevante |
| P2 Performance | ✅/❌ | Solo si hay bundle o API; si es CLI simple, skip |
| P3 Architecture | ✅ | Siempre relevante (incluye API design) |
| P4 Quality | ✅ | Siempre relevante (incluye deps) |
| P5 Experience | ❌ | Solo si hay interfaz de usuario real |
| P6 Operations | ✅/❌ | Versión reducida: scripts, env, CI básico; skip si no hay infra |

## Qué se omite en modo simple

| Elemento | Completo | Simple |
|----------|----------|--------|
| Matriz de superficie de ataque | Sí | No |
| Presupuesto de performance | Sí | No |
| Mapa de dependencias de módulos | Sí | No |
| Correlación profunda de causas raíz | Sí | Solo directa |
| Breaking Change Risk Map | Sí | No |
| Risk Score de dependencias | Sí | Versión reducida |
| Documentos por fase separados | Sí | Un solo archivo |
| Plan de tareas detallado | Sí | Solo bloqueantes |
| Scorecard con tendencia | Sí | Score actual nada más |

## Formato de output: `sequoia-simple.md`

```markdown
# Sequoia Simple Audit — [Proyecto]

**Fecha**: [fecha]
**Stack**: [detectado]
**Madurez**: [del mapa]

## Health Score
| Fase | Score |
|------|-------|
| Security | {emoji} |
| Architecture | {emoji} |
| Quality | {emoji} |
| Operations | {emoji} |
| **Global** | **{emoji}** |

## Hallazgos bloqueantes 🔴
{solo críticos, formato estándar de hallazgo}

## Hallazgos de riesgo 🟠
{solo riesgos altos}

## Quick wins
{3-5 acciones de mayor impacto con menor esfuerzo}

## Próximos pasos
1. [acción concreta más urgente]
2. [segunda más urgente]
3. [tercera]

---
Generado por Sequoia (flujo simple)
```

## Fast path: init + audit en uno

Para proyectos simples, se puede ejecutar init + audit en secuencia directa:

```bash
/sequoia init    # detecta proyecto simple
/sequoia audit   # auto-selecciona flujo simple
```

Si el init detecta que es simple, el audit sugiere automáticamente `--mode=quick`.

## Cuándo escalar a flujo completo

El flujo simple recomienda escalar cuando:
- El proyecto creció más de 50 archivos desde el init
- Se agregó CI/CD o infraestructura de deployment
- Se agregó base de datos o API pública
- El equipo creció a más de un desarrollador
- El proyecto pasó de prototipo a producción

En esos casos, el reporte simple incluye una nota:
> 📈 Este proyecto ha crecido. Considerá ejecutar `/sequoia audit --mode=full` para un análisis completo.
