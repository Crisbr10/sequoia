# Formato Estándar de Hallazgo

Formato obligatorio para TODOS los agentes de Sequoia. Sin excepciones.

## Plantilla markdown

```markdown
### [FASE-ID] · [Título del hallazgo]  [🔴 CRÍTICO | 🟠 RIESGO | 🟡 ATENCIÓN]

**Estado**: Confirmado | Parcial | No verificable | Desactualizado

**Evidencia**:
- `path/real/al/archivo.ext:línea` — descripción de lo observado
- Comportamiento o ausencia detectada

**Problema**:
Qué está mal y por qué técnicamente importa. Sin generalidades.

**Impacto real**:
Qué puede pasar en producción si esto sigue así.

**Recomendación mínima de alto leverage**:
Qué cambio concreto conviene hacer primero y por qué ese específicamente.

**Dependencias / bloqueos**:
Backend, infra, contrato de API, otros módulos, equipo externo, etc.

**Riesgo de implementación**: Bajo | Medio | Alto
Motivo del riesgo estimado.

**Criterio de aceptación**:
Cómo verificar que el hallazgo fue resuelto.
```

## Guía de cada campo

### ID del hallazgo

Formato: `[AGENT-ID}-{NNN}]`

- `C0` = Context, `P1` = Security, `P2` = Performance, `P3` = Architecture, `P4` = Quality, `P5` = Experience, `P6` = Operations
- `M1` = Correlator, `M2` = Reporter
- `NNN` = número secuencial dentro de la fase
- Ejemplo: `[P1-001]`, `[P3-012]`, `[P4-003]`

### Severidad

| Nivel | Emoji | Cuándo usarlo |
|-------|-------|---------------|
| CRÍTICO | 🔴 | Bloqueante de producción, seguridad explotable, pérdida de datos |
| RIESGO | 🟠 | Problema serio sin solución activa, degradación probable bajo carga |
| ATENCIÓN | 🟡 | Deuda técnica priorizable, mejora de calidad, future-proofing |

**Regla**: si dudás entre dos niveles, usá el menor. Es mejor sub-estimar que generar fatiga de alertas.

### Estado

| Estado | Significado |
|--------|-------------|
| Confirmado | Verificado contra código real, evidencia sólida |
| Parcial | Verificado parcialmente, algún aspecto no verificable |
| No verificable | Requiere acceso externo (infra, DB producción, logs) |
| Desactualizado | El hallazgo fue válido pero el código cambió desde entonces |

**Marcadores adicionales**:
- `[REQUIERE ACCESO EXTERNO]` — cuando no se puede verificar sin acceso a infra/logs
- `[SOLO SI ESCALA]` — cuando la recomendación solo aplica si el proyecto crece

### Evidencia

**Obligatorio** citar archivos reales con número de línea:

```markdown
**Evidencia**:
- `src/auth/handler.ts:45` — token almacenado en localStorage sin expiración
- `src/api/middleware.ts:12` — middleware de auth no valida refresh token
- Ausencia detectada: no existe archivo `.env.example`
```

Si no hay archivo (ausencia detectada), declararlo explícitamente.

### Problema

Descripción técnica del problema, no genérica.

❌ Mal: "El manejo de tokens no es seguro."
✅ Bien: "El JWT se almacena en localStorage, accesible por XSS. No hay mecanismo de revocación ni refresh token rotation."

### Impacto real

Qué pasa en producción, no en teoría.

❌ Mal: "Podría haber problemas de seguridad."
✅ Bien: "Un ataque XSS puede robar el token de sesión. No hay forma de invalidar sesiones comprometidas sin cambiar el secret."

### Recomendación

Una sola acción de mayor impacto. No una lista de 5 cosas.

❌ Mal: "Mejorar la seguridad de autenticación."
✅ Bien: "Mover el token a httpOnly cookie y agregar refresh token rotation. Ver `src/auth/handler.ts:45`."

### Riesgo de implementación

| Nivel | Cuándo |
|-------|--------|
| Bajo | Cambio localizado, sin dependencias, fácil rollback |
| Medio | Cambio en interfaz o contrato, requiere coordinación |
| Alto | Cambio en flujo central, afecta múltiples módulos, rollback complejo |

### Criterio de aceptación

Verificable y concreto. No "mejorar X".

```markdown
**Criterio de aceptación**:
- [ ] El token ya no se almacena en localStorage
- [ ] El token se envía en httpOnly cookie con flags Secure + SameSite=Strict
- [ ] Existe endpoint de refresh token con rotation
- [ ] Test E2E verifica que el token no es accesible por JS
```
