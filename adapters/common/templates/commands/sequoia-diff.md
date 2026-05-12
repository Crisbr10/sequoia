---
description: "Compara estado actual del proyecto contra la última auditoría registrada. Muestra: resuelto, nuevo, empeorado, sin cambio. Útil para tracking de evolución del proyecto."
allowed-tools: Read, Glob, Grep
---

# /sequoia diff

Compara el estado actual del proyecto contra la última auditoría registrada en Engram. Muestra evolución: qué mejoró, qué empeoró, qué apareció nuevo.

## Precondición

Debe existir al menos una auditoría previa en Engram. Si no hay auditoría previa, sugerir ejecutar `/sequoia audit` primero.

## Qué hace

1. Recupera la última auditoría de Engram
2. Ejecuta un scan rápido del estado actual del proyecto
3. Compara hallazgos anteriores vs estado actual
4. Clasifica cada hallazgo en una categoría de evolución
5. Genera el reporte de diff

## Categorías de comparación

| Categoría | Significado | Icono |
|-----------|-------------|-------|
| **Resuelto** | El hallazgo anterior ya no se reproduce | ✅ |
| **Nuevo** | Problema que no existía en la auditoría anterior | 🆕 |
| **Empeorado** | El hallazgo anterior sigue y ha worsened | 🔻 |
| **Sin cambio** | El hallazgo anterior sigue igual | ⏸️ |
| **Parcialmente resuelto** | Se mejoró pero no cumple criterio de aceptación | 🔸 |

## Flujo de ejecución

```
/sequoia diff
  │
  ├─ 1. Recuperar última auditoría de Engram
  │     ├─ Hallazgos con timestamp
  │     ├─ Health scores
  │     └─ Project Map snapshot
  │
  ├─ 2. Verificar cambios en estructura del proyecto
  │     ├─ ¿Archivos nuevos o eliminados desde la última auditoría?
  │     ├─ ¿Cambió el stack o las dependencias?
  │     └─ ¿Cambió la madurez del proyecto?
  │
  ├─ 3. Re-verificar cada hallazgo anterior
  │     ├─ Para cada hallazgo, leer los archivos citados
  │     ├─ ¿La evidencia sigue presente?
  │     ├─ ¿Se implementó la recomendación?
  │     └─ Clasificar: resuelto | sin cambio | empeorado | parcial
  │
  ├─ 4. Detectar hallazgos nuevos
  │     ├─ Scan rápido de áreas no cubiertas antes
  │     ├─ Solo hallazgos 🔴 y 🟠 (no es auditoría completa)
  │     └─ Listar como "nuevos"
  │
  └─ 5. Generar reporte de evolución
        ├─ Tabla resumen por categoría
        ├─ Comparativa de health scores
        └─ Persistir resultado en Engram
```

## Metodología de verificación

Para cada hallazgo anterior:

1. **Leer el archivo citado en la evidencia** — ¿existe aún? ¿tiene las mismas líneas?
2. **Verificar el criterio de aceptación** — ¿se cumplió?
3. **Cross-check con git blame/log** — ¿hubo commits que tocaron esa área?

Clasificación:
- Si el archivo cambió y el problema ya no está → ✅ Resuelto
- Si el archivo cambió pero el problema persiste parcialmente → 🔸 Parcial
- Si el archivo no cambió → ⏸️ Sin cambio
- Si el archivo cambió y hay problemas adicionales → 🔻 Empeorado

## Formato de salida

```markdown
## Sequoia Diff — [Proyecto]

**Auditoría anterior**: [fecha]
**Comparación actual**: [fecha]
**Tiempo transcurrido**: [días/semanas]

### Resumen de evolución

| Categoría | Cantidad | Porcentaje |
|-----------|----------|------------|
| ✅ Resueltos | {N} | {N}% |
| 🔸 Parciales | {N} | {N}% |
| ⏸️ Sin cambio | {N} | {N}% |
| 🔻 Empeorados | {N} | {N}% |
| 🆕 Nuevos | {N} | {N}% |
| **Total** | **{N}** | **100%** |

### Health Score comparativo

| Fase | Score anterior | Score actual | Tendencia |
|------|---------------|--------------|-----------|
| Security | 🟠 | 🟢 | ↗️ Mejorando |
| Performance | 🟡 | 🟡 | → Estable |
| ... | | | |

### Detalle de hallazgos resueltos ✅
{lista de hallazgos con qué cambió}

### Detalle de hallazgos nuevos 🆕
{solo hallazgos 🔴 y 🟠 detectados en el scan rápido}

### Detalle de hallazgos empeorados 🔻
{hallazgos donde el problema creció o se agregaron nuevos riesgos}

### Tendencia global
📈 Mejorando | ➡️ Estable | 📉 Degradando

### Recomendación
{cuándo ejecutar la próxima auditoría completa}
```

## Cuándo usar diff vs auditoría nueva

| Situación | Usar |
|-----------|------|
| Implementaste fixes y querés verificar | `diff` |
| Pasó 1-2 semanas y querés tracking | `diff` |
| Cambios grandes en el proyecto | `audit` (nueva auditoría) |
| Pasó más de un mes | `audit` (nueva auditoría) |
| Nuevo miembro en el equipo | `audit` (nueva auditoría) |
| Post-merge de feature grande | `diff` primero, `audit` si hay sorpresas |

## Detección de obsolescencia

Si la última auditoría tiene más de 30 días, diff muestra una advertencia:
> ⚠️ La última auditoría tiene {N} días. Los hallazgos pueden estar desactualizados. Considerá ejecutar `/sequoia audit` para una auditoría fresca.

Si el Project Map cambió significativamente (nuevos deps, cambio de framework, etc.), diff recomienda ejecutar un `init` + `audit` nuevo.
