# Flujo: Auditoría Incremental

Flujo para re-auditoría y tracking de evolución del proyecto.

## Cuándo usar

| Situación | Comando |
|-----------|---------|
| Después de implementar fixes de auditoría previa | `/sequoia diff` |
| Health check periódico semanal/quincenal | `/sequoia diff` |
| Post-merge de feature grande | `/sequoia diff` → si hay sorpresas, `audit` |
| Cambios significativos en el proyecto | `/sequoia audit` (nueva completa) |
| Más de 30 días desde la última auditoría | `/sequoia audit` (nueva completa) |

## Flujo de diff incremental

```
/sequoia diff
  │
  ├─ 1. RECUPERAR AUDITORÍA ANTERIOR
  │     ├─ Hallazgos de Engram (más reciente)
  │     ├─ Health scores
  │     └─ Snapshot del estado (commit hash, estructura)
  │
  ├─ 2. DETECTAR STALENESS
  │     ├─ ¿Cuántos commits desde la última auditoría?
  │     ├─ ¿Cuántos archivos cambiaron?
  │     └─ ¿Cambió el stack o estructura significativamente?
  │
  ├─ 3. RE-VERIFICAR HALLAZGOS ANTERIORES
  │     ├─ Para cada hallazgo previo:
  │     │   ├─ Leer archivos citados en la evidencia
  │     │   ├─ Comparar estado actual vs snapshot
  │     │   └─ Clasificar: ✅ | 🔸 | ⏸️ | 🔻
  │     └─ Generar tabla de clasificación
  │
  ├─ 4. SCAN RÁPIDO DE HALLAZGOS NUEVOS
  │     ├─ Solo en áreas cambiadas desde la última auditoría
  │     ├─ Solo 🔴 CRÍTICO y 🟠 RIESGO
  │     └─ No es auditoría completa: es barrido rápido
  │
  ├─ 5. CALCULAR EVOLUCIÓN
  │     ├─ Score anterior vs score actual (estimado)
  │     ├─ Tendencia: ↗️ Mejorando | → Estable | ↘️ Degradando
  │     └─ Velocidad de resolución (hallazgos resueltos / tiempo)
  │
  └─ 6. GENERAR REPORTE DE EVOLUCIÓN
        └─ Formato diff (ver sequoia-diff.md)
```

## Detección de staleness

```markdown
| Indicador | Verde | Amarillo | Rojo |
|-----------|-------|----------|------|
| Días desde última auditoría | < 14 | 14-30 | > 30 |
| Commits desde última auditoría | < 20 | 20-50 | > 50 |
| Archivos cambiados | < 15% | 15-40% | > 40% |
| Cambios en deps | 0 | 1-3 | > 3 |
| Cambio de estructura | No | Menor | Significativo |
```

- **Todo verde**: diff incremental es suficiente
- **Alguna amarilla**: diff + atención a esas áreas
- **Alguna roja**: recomendar auditoría completa nueva

## Scope incremental

El scan rápido solo re-audita áreas que cambiaron:

1. **Obtener diff de archivos** desde el commit de la última auditoría
2. **Filtrar** a archivos de código fuente (excluir generated, vendor, lockfiles)
3. **Para cada archivo cambiado**, ejecutar solo los agentes relevantes al tipo de archivo
4. **No re-ejecutar** agentes sobre archivos sin cambios

Esto reduce el tiempo de scan de ~15-30 min a ~3-8 min.

## Scoring de evolución

### Score por fase

Comparar score anterior con estimación actual:

```
🟢 → 🟢 = → Estable (mantiene salud)
🟡 → 🟢 = ↗️ Mejorando (resolvió deuda)
🟠 → 🟢 = ↗️↗️ Mejora significativa
🟠 → 🟡 = ↗️ Mejorando
🟢 → 🟡 = ↘️ Degradando levemente
🟡 → 🟠 = ↘️ Degradando
🟢 → 🟠 = ↘️↘️ Degradación significativa
🟢 → 🔴 = 🔻 Crítico (requiere acción inmediata)
```

### Trend global

```
Improvement rate = (resueltos + parciales) / total_hallazgos_previos

📈 Improving:  rate > 30%
➡️ Stable:     rate 10-30%
📉 Degrading:  rate < 10%  O  nuevos > resueltos
```

### Velocity score

```markdown
| Métrica | Fórmula | Interpretación |
|---------|---------|----------------|
| Resolución rate | resueltos / hallazgos_previos | % de progreso |
| Nueva deuda rate | nuevos / semanas_transcurridas | velocidad de aparición |
| Net trend | (resueltos - nuevos) / semanas | balance neto |
```

## Cuándo auditar completo vs incremental

```
                 ┌──────────────────────────────┐
                 │  ¿Cuánto cambió desde la      │
                 │  última auditoría?             │
                 └──────────┬───────────────────┘
                            │
                 ┌──────────▼───────────────────┐
                 │  ¿Cambió el stack o la        │
              ┌──┤  estructura significativamente?│
              │  └──────────┬───────────────────┘
              │             │
         Sí   │        No   │
              │             │
    ┌─────────▼──┐   ┌─────▼──────────┐
    │ AUDITORÍA  │   │ ¿Staleness     │
    │ COMPLETA   │   │ rojo?          │
    │ NUEVA      │   └───┬──────┬─────┘
    └────────────┘       │      │
                     Sí  │   No │
                    ┌────▼──┐ ┌─▼──────────┐
                    |AUDIT  │ │ DIFF        │
                    |COMPLETO│ │ INCREMENTAL │
                    └───────┘ └─────────────┘
```

## Persistencia del diff

Cada diff se guarda en Engram con:
- **title**: "Sequoia Diff — {proyecto} — {fecha}"
- **topic_key**: `sequoia/{proyecto}/diff-{timestamp}`
- **type**: `architecture`
- **content**: resultado completo del diff

Esto permite construir un historial de evolución. El scorecard puede mostrar tendencias a lo largo de múltiples diffs.

## Integración con audit completo

Los diffs NO reemplazan auditorías completas. Son complementarios:

- **Audit completo**: baseline, descubrimiento exhaustivo, correlación profunda
- **Diff incremental**: tracking, verificación de fixes, detección temprana de degradación

Cadencia sugerida:
- Audit completo: mensual o ante cambios grandes
- Diff incremental: semanal o post-fix
