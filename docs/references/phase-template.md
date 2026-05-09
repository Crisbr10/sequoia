# Plantilla Obligatoria por Fase

Estructura que TODO documento de fase generado por `sequoia-reporter` debe seguir. Sin excepciones.

## Plantilla

```markdown
# Fase [N] — [Nombre]

**Agente**: sequoia-[nombre]
**Proyecto**: [nombre]
**Fecha de auditoría**: [fecha]
**Stack detectado**: [del mapa de proyecto]
**Modo**: Full | Quick | Review

---

## 1. Objetivo de la fase

Qué dominio cubre esta fase y qué se busca evaluar. 2-3 líneas.

## 2. Scope de inspección

### Archivos y directorios revisados
- `{path/}` — [motivo de inclusión]
- `{path/}` — [motivo de inclusión]

### Qué quedó fuera del scope
- `{path/}` — [motivo de exclusión: generado, vendor, fuera de contexto]
- `{path/}` — [motivo de exclusión]

## 3. Estado actual verificado

### Qué dice la documentación interna (si existe)
[Resumen de docs relevantes. Si no hay docs: "No se encontró documentación interna relevante."]

### Qué se confirmó en código
[Hechos verificados contra código real.]

### Qué quedó desactualizado, ambiguo o no verificable
- [Si hay docs que contradicen el código, listar aquí]
- [Si hay aspectos que no se pudieron verificar, declarar con `[NO VERIFICABLE]` o `[REQUIERE ACCESO EXTERNO]`]

## 4. Hallazgos consolidados

Ordenados por severidad (🔴 → 🟠 → 🟡). Sin duplicados. Sin relleno.

### [{AGENT}-{NNN}] · [Título del hallazgo]  [🔴 CRÍTICO | 🟠 RIESGO | 🟡 ATENCIÓN]

**Estado**: Confirmado | Parcial | No verificable | Desactualizado

**Evidencia**:
- `path/real/al/archivo.ext:línea` — descripción

**Problema**:
[Descripción técnica concreta]

**Impacto real**:
[Qué pasa en producción]

**Recomendación mínima de alto leverage**:
[Una sola acción de máximo impacto]

**Dependencias / bloqueos**:
[Si aplica]

**Riesgo de implementación**: Bajo | Medio | Alto
[Motivo]

**Criterio de aceptación**:
- [ ] [condición verificable]

---

*(Repetir para cada hallazgo)*

## 5. Faltantes de alto leverage

Solo mejoras justificadas técnicamente. Con impacto esperado. No wishlist.

| Mejora | Impacto esperado | Esfuerzo estimado |
|--------|-----------------|-------------------|
| [descripción] | [qué cambia] | [bajo/medio/alto] |

## 6. Plan de tareas

### [TASK-{N}.{NNN}] · [Título de la tarea]

- **Contexto**: [por qué se necesita, 2-3 líneas]
- **Archivos/módulos candidatos**: `path/to/file`
- **Impacto en la app actual**: [qué mejora]
- **Dependencia previa**: [TASK-XX.XXX si aplica, o "ninguna"]
- **Riesgo de rotura**: Bajo | Medio | Alto — [motivo]
- **Criterio de aceptación**:
  - [ ] [condición verificable]
- **Prioridad**: 🔴 Bloqueante | 🟠 Alto leverage | 🟡 Backlog

---

*(Repetir para cada tarea)*

## 7. Orden de implementación recomendado

Secuencia que minimiza riesgo y maximiza impacto:

```
1. [TASK-XX.001] → causa raíz, desbloquea otras tareas
2. [TASK-XX.003] → dependiente de 001, quick win
3. [TASK-XX.002] → independiente, se puede hacer en paralelo
4. [TASK-XX.004] → requiere que 001 y 003 estén completas
```

Justificación del orden: [por qué esta secuencia]

## 8. Riesgos y bloqueos de la fase

| Riesgo/Bloqueo | Impacto | Mitigación |
|---------------|---------|------------|
| [descripción] | [alto/medio/bajo] | [acción sugerida] |

## 9. Checklist de cierre de fase

Lista verificable: qué debe ser verdad cuando esta fase esté "done".

- [ ] Todos los hallazgos 🔴 tienen tarea asignada
- [ ] Todas las tareas tienen criterio de aceptación
- [ ] Las dependencias entre tareas están declaradas
- [ ] El orden de implementación está justificado
- [ ] Los hallazgos no verificables están marcados como tales
- [ ] No hay recomendaciones genéricas sin evidencia específica
- [ ] Los archivos citados en la evidencia existen y son correctos
```

## Reglas de la plantilla

1. **No secciones vacías**. Si una sección no aplica, escribir "No aplica: [motivo]" en vez de dejarla en blanco.
2. **Todo hallazgo usa el formato estándar** de `references/finding-format.md`.
3. **Cada tarea es autosuficiente**: no requiere leer otras secciones para entenderse.
4. **Las secciones 1-3 son contexto**, las 4-6 son accionables, las 7-9 son de cierre.
5. **El checklist de cierre** debe ser verificable sin ambigüedad.
