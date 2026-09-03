

# II GONGCHACUP 2026

Reglamento oficial de la competencia tecnológica

###### **Edición: Sistema de Gestión de Bebidas**

|                  |                             |
|------------------|-----------------------------|
| <b>Fechas</b>    | 3 y 4 de septiembre de 2026 |
| <b>Modalidad</b> | Semipresencial              |
| <b>Equipos</b>   | Máximo 2 integrantes        |

II Congreso de Tecnologías en Ciencias Computacionales — CONTECS 2026

## 1. Presentación

La II GongchaCUP 2026 es una competencia de desarrollo de software que reta a estudiantes universitarios a resolver un problema real del sector de bebidas. Durante dos días, los equipos deberán analizar el dominio, modelar recetas encadenadas y construir una aplicación web funcional que conecte inventario, producción y planificación.

## 2. Problemática

Los establecimientos que preparan bebidas manejan materias primas, productos parcialmente procesados y bebidas finales. Cuando estos elementos se controlan de forma separada o manual, resulta difícil conocer la capacidad real de producción, anticipar faltantes, aprovechar excedentes y calcular compras para pedidos o eventos.

La solución deberá responder, como mínimo, las siguientes preguntas:

- ¿Cuántas unidades de cada bebida pueden elaborarse con el inventario disponible?
- ¿Qué cantidades de insumos y productos semiterminados se requieren para producir una cantidad solicitada?
- ¿Qué existencias quedarían después de una producción y cuál es el insumo crítico?
- ¿Qué cantidades consolidadas se necesitan para atender un evento con varias bebidas?

## 3. Objetivo del reto

Desarrollar una aplicación web intuitiva y confiable que modele el proceso de preparación de bebidas y calcule la producción en doble vía: desde el inventario hacia la producción posible y desde una demanda hacia los recursos requeridos.

## 4. Modelo del dominio

| Nivel              | Definición                                                                            | Ejemplo                                      |
|--------------------|---------------------------------------------------------------------------------------|----------------------------------------------|
| Materia prima      | Insumo disponible en su estado de compra y almacenado en una unidad definida.         | Azúcar, leche, té seco, tapioca cruda.       |
| Semiterminado      | Preparación intermedia obtenida mediante una receta y utilizada por otras recetas.    | Jarabe simple, té preparado, tapioca cocida. |
| Producto terminado | Bebida final entregada al cliente y compuesta por materias primas y/o semiterminados. | Té con leche y tapioca.                      |

Una receta puede depender de otra receta. El sistema deberá resolver estas dependencias sin contabilizar dos veces un mismo consumo y deberá impedir ciclos de producción inválidos.

## 5. Alcance funcional obligatorio

### 5.1 Catálogo e inventario

- Crear, consultar, editar y eliminar productos.
- Registrar código, nombre, descripción, tipo, unidad de medida, existencia y una imagen o referencia de imagen.
- Distinguir materias primas, semiterminados y productos terminados.
- Validar cantidades no negativas y unidades compatibles.

### 5.2 Gestión de recetas

- Crear y modificar recetas indicando producto resultante, rendimiento del lote, componentes y cantidades.
- Permitir que una receta utilice materias primas y productos semiterminados.
- Visualizar la composición completa de una bebida, incluyendo dependencias intermedias.
- Detectar recetas incompletas, referencias circulares o componentes inexistentes.

### 5.3 Motor de cálculo en doble vía

- Cálculo directo: determinar cuántas unidades completas de una bebida pueden producirse con el inventario actual.
- Cálculo inverso: determinar los insumos requeridos para una cantidad solicitada de bebidas.
- Mostrar el detalle de consumo, los excedentes y el insumo limitante o crítico.
- No redondear hacia arriba la producción posible: solo cuentan unidades completas.
- Permitir simular cálculos sin alterar existencias y ejecutar una producción confirmada que sí descuente inventario.

### 5.4 Planificación de eventos

El sistema deberá recibir una demanda compuesta por varias bebidas y consolidar el total de materias primas y semiterminados necesarios. Debe identificar faltantes frente al inventario actual y evitar duplicar componentes compartidos.

### 5.5 Interfaz y trazabilidad

- Navegación clara y diseño adaptable a computadoras.
- Mensajes comprensibles ante errores y validaciones.
- Resumen de cada cálculo con fecha, productos, cantidades y resultado.
- Consulta del inventario antes y después de una producción confirmada.

## 6. Tecnología y restricciones técnicas

El stack tecnológico es libre. Cada equipo podrá seleccionar lenguajes, frameworks, bibliotecas y motor de base de datos, siempre que cumpla las siguientes condiciones:

- La solución debe ser una aplicación web y utilizar una base de datos persistente (local, servidor o nube).
- Debe poder ejecutarse y demostrarse durante la evaluación. Si depende de Internet, el equipo debe contar con un plan de contingencia local.
- El código deberá mantenerse en el repositorio GitHub asignado por el Comité Organizador del CONTECS 2026 desde el inicio de la competencia.
- Se permiten bibliotecas, frameworks, documentación pública y asistentes de inteligencia artificial. Su uso deberá declararse en el documento técnico y mostrar el dominio de entendimiento de lo que no ha sido creado por cuenta propia, de lo contrario se verá afectada su calificación final.
- No se permite reutilizar una aplicación previamente construida ni recibir desarrollo directo de personas ajenas al equipo.

## 7. Datos oficiales del reto

La organización entregará el archivo “Datos\_Prueba\_GongchaCUP\_2026.xlsx”, que contiene catálogo de productos, inventario inicial, recetas, componentes, demanda de evento y casos de validación. Los datos son ficticios y se proporcionan exclusivamente con fines académicos y competitivos.

- Los equipos podrán importar los datos, cargarlos mediante scripts o registrarlos desde la interfaz.
- La solución debe conservar las unidades indicadas y documentar cualquier conversión.
- El jurado podrá utilizar datos adicionales o cantidades diferentes para verificar que los cálculos sean dinámicos y no estén codificados de forma fija.

## 8. Participantes e inscripción

- Podrán participar estudiantes universitarios inscritos y aceptados por la organización.
- Cada equipo estará integrado por un máximo de dos personas.
- La inscripción cierra el miércoles 2 de septiembre de 2026 a las 12:00 m.d.
- No se permitirán cambios de integrantes después del inicio, salvo autorización expresa de la organización.
- Cada participante deberá asistir al acto de inauguración, a los bloques presenciales obligatorios, a la presentación final y a la premiación.

## 9. Cronograma oficial

| Fecha y hora                           | Actividad                                                      | Modalidad              |
|----------------------------------------|----------------------------------------------------------------|------------------------|
| Miércoles 2 de septiembre — 12:00 m.d. | Cierre de inscripciones                                        | En línea               |
| Jueves 3 — 9:00 a.m. a 10:30 a.m.      | Inauguración, presentación del reto, reglas y entrega de datos | Presencial obligatoria |
| Jueves 3 — 10:30 a.m. a 3:00           | Desarrollo acompañado y apertura del repositorio oficial       | Presencial             |

| Fecha y hora                               | Actividad                              | Modalidad              |
|--------------------------------------------|----------------------------------------|------------------------|
| p.m.                                       |                                        |                        |
| Jueves 3, 3:00 p.m. a viernes 4, 8:00 a.m. | Continuación del desarrollo            | Remota                 |
| Viernes 4 — 8:00 a.m. a 10:30 a.m.         | Desarrollo final y entrega             | Presencial obligatoria |
| Viernes 4 — 10:30 a.m. a 12:00 m.d.        | Preparación de presentaciones y receso | Presencial             |
| Viernes 4 — 12:00 m.d. a 2:00 p.m.         | Presentaciones y evaluación del jurado | Presencial obligatoria |
| Viernes 4 — 2:00 p.m. a 3:00 p.m.          | Deliberación del jurado                | Presencial             |
| Viernes 4 — 3:00 p.m.                      | Premiación y acto de cierre            | Presencial obligatoria |

La hora oficial será la indicada por la organización. Los cambios necesarios por causas logísticas serán comunicados a todos los equipos por el canal oficial.

## 10. Entregables

Antes de las 10:30 a.m. del viernes 4 de septiembre, cada equipo deberá dejar disponible en el repositorio oficial:

- Código fuente completo y ejecutable.
- Archivo o scripts de creación y carga de la base de datos.
- README con requisitos, instalación, configuración, ejecución y credenciales de demostración.
- Documento técnico breve en PDF: arquitectura, modelo de datos, decisiones, algoritmo de cálculo, pruebas, limitaciones y declaración de herramientas de IA utilizadas.
- Aplicación funcional preparada para demostración.
- Presentación final de máximo 7 minutos, seguida de hasta 5 minutos de preguntas del jurado.

Solo se evaluará la versión registrada en el repositorio al momento del cierre. El Comité Organizador podrá solicitar una copia ejecutable o respaldo adicional.

## 11. Evaluación

| Criterio                        | Puntos | Evidencia esperada                                                                                                                         |
|---------------------------------|--------|--------------------------------------------------------------------------------------------------------------------------------------------|
| Exactitud funcional y cobertura | 30     | CRUD; recetas encadenadas; cálculo directo e inverso; excedentes; insumo crítico; evento; actualización de inventario y manejo de errores. |
| Diseño técnico y datos          | 15     | Modelo coherente, persistencia, arquitectura, seguridad básica,                                                                            |

| Criterio                         | Puntos | Evidencia esperada                                                                            |
|----------------------------------|--------|-----------------------------------------------------------------------------------------------|
|                                  |        | mantenibilidad y algoritmo generalizable.                                                     |
| Experiencia de usuario           | 20     | Interfaz clara, navegación, mensajes, legibilidad y visualización útil de resultados.         |
| Calidad, pruebas y documentación | 10     | Pruebas relevantes, instrucciones reproducibles, código organizado y decisiones documentadas. |
| Presentación y defensa           | 25     | Demostración completa, comunicación, dominio técnico y respuestas al jurado.                  |

Puntaje total: 100 puntos. Los criterios obligatorios prevalecen sobre características adicionales.

## 12. Desempates y bonificaciones

En caso de empate, se decidirá en este orden: (1) mayor puntaje en exactitud funcional, (2) mayor puntaje en diseño técnico, (3) mejor resultado en una prueba sorpresa y (4) decisión razonada del jurado.

Las funcionalidades adicionales solo aportarán dentro de los criterios existentes y no compensarán errores en los cálculos obligatorios. Ejemplos: alertas de stock, historial, exportación de reports, panel visual o pruebas automatizadas.

## 13. Normas de conducta, autoría y propiedad

- Los participantes deberán actuar con respeto hacia otros equipos, jurado, patrocinadores y personal organizador.
- El trabajo presentado debe haber sido desarrollado por los integrantes durante el periodo oficial de la competencia.
- Todo recurso externo significativo deberá atribuirse. El plagio, la suplantación, la alteración de repositorios o el acceso no autorizado a soluciones ajenas implicarán descalificación.
- Cada equipo conserva la autoría de su solución. La organización podrá tomar fotografías, grabar las presentaciones y difundir resultados con fines académicos y promocionales, reconociendo a sus autores.
- Los datos de la competencia no representan recetas comerciales reales de Gong cha ni información confidencial de patrocinadores.

## 14. Penalizaciones y descalificación

| Situación                   | Medida                                                                                     |
|-----------------------------|--------------------------------------------------------------------------------------------|
| Entrega posterior al cierre | La versión tardía no será considerada, salvo falla general confirmada por la organización. |

| Situación                                      | Medida                                                                                  |
|------------------------------------------------|-----------------------------------------------------------------------------------------|
| Ausencia en una actividad obligatoria          | Podrá impedir la evaluación o causar descalificación, según la situación.               |
| Aplicación no ejecutable                       | Se evaluará únicamente la evidencia verificable disponible; no se presumirán funciones. |
| Datos o cálculos codificados para casos fijos  | Pérdida de puntos en exactitud y diseño técnico.                                        |
| Plagio, ayuda externa directa o conducta grave | Descalificación del equipo.                                                             |

## 15. Jurado y decisiones

El jurado estará conformado por profesionales designados por la organización. Podrá inspeccionar el código, modificar datos de prueba, solicitar la repetición de cálculos y formular preguntas técnicas. Sus decisiones serán definitivas. Cualquier situación no contemplada será resuelta por el Comité Organizador de la GongchaCUP 2026.

## 16. Premios previstos

- Primer lugar: dos monitores LG UltraGear.
- Segundo lugar: dos Redmi Watch 5 Active y dos Redmi Buds 6 Play.
- Tercer lugar: dos mouse Corsair.

Los premios están sujetos a confirmación final, disponibilidad y condiciones comunicadas por el Comité Organizador de CONTECS 2026. No son transferibles ni canjeables por efectivo, salvo disposición expresa.

## 17. Aceptación del reglamento

La inscripción y participación implican la lectura y aceptación integral de este reglamento, de las comunicaciones oficiales y de las decisiones adoptadas por la organización dentro del marco de la competencia.

## Anexo A. Criterios mínimos de aceptación funcional

| ID | Criterio                                                                    |
|----|-----------------------------------------------------------------------------|
| A1 | El sistema registra los tres tipos de producto y sus unidades.              |
| A2 | Una receta puede consumir otra preparación semiterminada.                   |
| A3 | El cálculo directo limita la producción por el componente realmente escaso. |

| ID  | Criterio                                                                       |
|-----|--------------------------------------------------------------------------------|
| A4  | El cálculo inverso descompone la demanda hasta los insumos requeridos.         |
| A5  | Los excedentes se calculan después del consumo y nunca son negativos.          |
| A6  | La demanda de varias bebidas se consolida antes de compararse con inventario.  |
| A7  | Una simulación no altera el inventario; una producción confirmada sí.          |
| A8  | Los resultados cambian correctamente cuando cambian las existencias o recetas. |
| A9  | El sistema maneja recetas incompletas, ciclos y cantidades inválidas.          |
| A10 | La aplicación puede instalarse y ejecutarse siguiendo el README.               |