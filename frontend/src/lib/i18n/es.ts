/**
 * Spanish copy registry for Gong Cha inventory management system.
 * All user-facing text should come from this registry.
 */

export const es = {
  // ============================================================================
  // Global
  // ============================================================================
  app: {
    name: 'Gong Cha — Sistema de Gestión',
    skipToMain: 'Saltar al contenido principal',
  },

  // ============================================================================
  // Navigation
  // ============================================================================
  nav: {
    dashboard: 'Panel',
    catalog: 'Catálogo',
    inventory: 'Inventario',
    recipes: 'Recetas',
    calculations: 'Cálculos',
    events: 'Eventos',
    production: 'Producción',
  },

  // ============================================================================
  // Dashboard
  // ============================================================================
  dashboard: {
    title: 'Panel de Control',
    description: 'Resumen operativo y accesos rápidos',
    welcome: 'Bienvenido al sistema de gestión de Gong Cha',
    placeholder: 'Una vista breve para decidir qué revisar ahora.',
    kpiProducts: 'Productos activos',
    kpiStockLines: 'Líneas de inventario',
    kpiRecipes: 'Recetas disponibles',
    kpiLocations: 'Ubicaciones en uso',
    taskTitle: 'Siguiente acción',
    taskCapacity: 'Revisar capacidad',
    taskCapacityDescription: 'Calcula cuántas unidades completas permite el inventario actual.',
    taskDemand: 'Explorar requerimientos',
    taskDemandDescription: 'Expande una cantidad pedida hasta componentes y materias primas.',
    taskEvent: 'Planificar un evento',
    taskEventDescription: 'Consolida la demanda y detecta faltantes antes de comprar.',
    taskProduction: 'Preparar producción',
    taskProductionDescription: 'Simula el consumo y confirma el descuento solo cuando estés listo.',
    openTask: 'Abrir formulario',
    refresh: 'Actualizar indicadores',
  },

  // ============================================================================
  // API Status
  // ============================================================================
  apiStatus: {
    checking: 'Verificando conexión...',
    connected: 'Conectado',
    disconnected: 'Sin conexión',
    error: 'Error de conexión',
    retry: 'Reintentar',
    failedToConnect: 'No se pudo conectar con la API',
    checkConnection: 'Verifica que el servidor esté activo en',
  },

  // ============================================================================
  // Errors
  // ============================================================================
  errors: {
    generic: 'Ocurrió un error inesperado',
    network: 'Error de red. Verifica tu conexión.',
    timeout: 'La solicitud tardó demasiado. Intenta nuevamente.',
    cancelled: 'La solicitud fue cancelada',
    invalidResponse: 'Respuesta inválida del servidor',
    notFound: 'Recurso no encontrado',
    unauthorized: 'No autorizado',
    forbidden: 'Acceso prohibido',
    conflict: 'Conflicto con el estado actual',
    validationError: 'Error de validación',
    serverError: 'Error del servidor',
    dependencyCycle: 'Ciclo de dependencia detectado',
    missingComponent: 'Falta un componente o el producto no existe',
    unitMismatch: 'La unidad no coincide con la del producto',
    duplicateComponent: 'La receta no puede repetir el mismo componente',
    selfComponent: 'Un producto no puede ser componente de su propia receta',
    insufficientInventory: 'Inventario insuficiente. No se descontaron existencias.',
  },

  // ============================================================================
  // Product types
  // ============================================================================
  productType: {
    materia_prima: 'Materia Prima',
    semiterminado: 'Semiterminado',
    producto_terminado: 'Producto Terminado',
    raw_material: 'Materia Prima',
    semi_finished: 'Semiterminado',
    finished_product: 'Producto Terminado',
  },

  // ============================================================================
  // Units
  // ============================================================================
  units: {
    kg: 'kg',
    l: 'L',
    ml: 'ml',
    g: 'g',
    unidad: 'unidad',
    unit: 'unidad',
    porcion: 'porción',
  },

  // ============================================================================
  // Common actions
  // ============================================================================
  actions: {
    save: 'Guardar',
    cancel: 'Cancelar',
    delete: 'Eliminar',
    edit: 'Editar',
    create: 'Crear',
    close: 'Cerrar',
    confirm: 'Confirmar',
    retry: 'Reintentar',
    back: 'Volver',
    next: 'Siguiente',
    previous: 'Anterior',
    search: 'Buscar',
    filter: 'Filtrar',
    clear: 'Limpiar',
    submit: 'Enviar',
    loading: 'Cargando...',
    add: 'Añadir',
    remove: 'Quitar',
    calculate: 'Calcular',
  },

  // ============================================================================
  // Forms
  // ============================================================================
  forms: {
    required: 'Este campo es obligatorio',
    invalidFormat: 'Formato inválido',
    mustBePositive: 'Debe ser un número positivo',
    mustBeGreaterThanZero: 'Debe ser mayor que cero',
    validationErrors: 'Hay errores en el formulario',
    fieldError: 'Error en el campo',
  },

  // ============================================================================
  // States
  // ============================================================================
  states: {
    empty: 'No hay datos para mostrar',
    loading: 'Cargando información...',
    error: 'Error al cargar los datos',
    noResults: 'No se encontraron resultados',
  },

  // ============================================================================
  // Catalog
  // ============================================================================
  catalog: {
    title: 'Catálogo de Productos',
    description: 'Gestión del catálogo completo',
    searchPlaceholder: 'Buscar por nombre o ID...',
    filterByType: 'Filtrar por tipo',
    allTypes: 'Todos los tipos',
    viewGrid: 'Vista de tarjetas',
    viewTable: 'Vista de tabla',
    productCount: 'productos',
    noProducts: 'No hay productos disponibles',
    noProductsFiltered: 'No se encontraron productos con los filtros aplicados',
  },

  // ============================================================================
  // Inventory
  // ============================================================================
  inventory: {
    title: 'Inventario',
    description: 'Existencias actuales por ubicación',
    searchPlaceholder: 'Buscar por producto o ubicación...',
    filterByLocation: 'Filtrar por ubicación',
    allLocations: 'Todas las ubicaciones',
    productId: 'ID Producto',
    quantity: 'Cantidad',
    location: 'Ubicación',
    lastUpdated: 'Última actualización',
    noInventory: 'No hay existencias registradas',
    noInventoryFiltered: 'No se encontraron existencias con los filtros aplicados',
    adjust: 'Ajustar',
    adjustInventory: 'Ajustar Inventario',
    setQuantity: 'Establecer Cantidad',
    adjustSuccess: 'Inventario ajustado exitosamente',
    quantityPlaceholder: 'Ej: 150.5',
    locationPlaceholder: 'Ej: Bodega principal',
  },

  // ============================================================================
  // Products
  // ============================================================================
  product: {
    id: 'ID',
    name: 'Nombre',
    type: 'Tipo',
    unit: 'Unidad',
    description: 'Descripción',
    imageRef: 'Imagen',
    noDescription: 'Sin descripción',
    noImage: 'Sin imagen',
    create: 'Crear Producto',
    edit: 'Editar Producto',
    createSuccess: 'Producto creado exitosamente',
    updateSuccess: 'Producto actualizado exitosamente',
    deleteSuccess: 'Producto eliminado exitosamente',
    productIdPlaceholder: 'Ej: TEA001',
    namePlaceholder: 'Ej: Té verde jazmín',
    descriptionPlaceholder: 'Descripción opcional del producto',
    imageRefPlaceholder: 'Ej: te_verde.jpg',
  },

  // ============================================================================
  // Accessibility
  // ============================================================================
  a11y: {
    menu: 'Menú',
    close: 'Cerrar',
    loading: 'Cargando',
    error: 'Error',
    success: 'Éxito',
    warning: 'Advertencia',
    info: 'Información',
    productImage: 'Imagen del producto',
    dismissError: 'Cerrar aviso de error',
  },

  // ============================================================================
  // Recipes
  // ============================================================================
  recipes: {
    title: 'Recetas',
    description: 'Composición y dependencias de producción',
    listLabel: 'Listado de recetas',
    create: 'Nueva receta',
    edit: 'Editar receta',
    empty: 'No hay recetas registradas',
    recipeId: 'ID de receta',
    resultProduct: 'Producto resultante',
    batchYield: 'Rendimiento del lote',
    yieldUnit: 'Unidad de rendimiento',
    components: 'Componentes',
    component: 'Componente',
    quantity: 'Cantidad',
    unit: 'Unidad',
    addComponent: 'Agregar componente',
    removeComponent: 'Quitar componente',
    composition: 'Árbol de composición',
    selectRecipe: 'Selecciona una receta para ver su composición',
    cycleBadge: 'Ciclo',
    saving: 'Guardando...',
    noMoreComponents: 'No hay más componentes disponibles',
    loadingList: 'Cargando recetas...',
    loadingDetail: 'Cargando composición...',
    loadError: 'No se pudieron cargar las recetas',
    createTitle: 'Crear receta',
    editTitle: 'Editar receta',
    closeEditor: 'Cerrar editor de receta',
    nestedRecipe: 'Incluye receta anidada',
    chooseProduct: 'Seleccionar producto',
    chooseComponent: 'Seleccionar componente',
    validationSummary: 'Corrige los errores del formulario',
    componentCount: 'componentes',
  },

  // ============================================================================
  // Calculations
  // ============================================================================
  calculations: {
    title: 'Cálculos',
    description: 'Capacidad directa a partir del inventario y requerimientos inversos por receta',
    directTitle: 'Capacidad directa',
    directDescription: 'Unidades completas que se pueden producir con el inventario actual o una instantánea sustituida',
    inverseTitle: 'Requerimientos inversos',
    inverseDescription: 'Componentes inmediatos y materias primas necesarias para una cantidad pedida',
    product: 'Producto',
    productId: 'ID',
    productPlaceholder: 'Seleccionar producto',
    quantity: 'Cantidad',
    results: 'Resultado',
    maxUnits: 'Unidades máximas',
    limitingComponent: 'Componente limitante',
    leftovers: 'Sobrantes',
    leftoversCaption: 'Sobrantes de inventario después de producir el máximo de unidades completas',
    immediate: 'Componentes inmediatos',
    immediateCaption: 'Componentes directos de la receta, sin expandir semiterminados',
    rawMaterials: 'Materias primas',
    rawMaterialsCaption: 'Requerimientos expandidos hasta materia prima',
    incomplete: 'Recetas incompletas',
    inventoryOverrides: 'Sustituciones de inventario',
    overrideHint: 'Cada fila reemplaza la existencia de un producto en la instantánea enviada al cálculo',
    addOverride: 'Añadir sustitución',
    removeOverride: 'Quitar sustitución',
    replaceInventory: 'Usar solo estas existencias (ignorar inventario oficial)',
    useDirect: 'Consumir semiterminado del inventario',
    useDirectHint: 'Los productos listados no se expanden a materia prima; se descuentan de existencias',
    addUseDirect: 'Añadir a uso directo',
    removeUseDirect: 'Quitar uso directo',
    noLeftovers: 'No hay sobrantes que mostrar',
    noRequirements: 'Sin requerimientos',
    leftoverOf: 'Quedan {current} de {max} {unit}',
    exhausted: 'Agotado',
    selectProduct: 'Selecciona un producto',
    invalidQuantity: 'Indica una cantidad mayor que cero',
    invalidOverride: 'Cada sustitución necesita producto y una cantidad numérica (cero o más)',
    unitsComplete: 'unidades completas',
    none: 'Ninguno',
  },

  // ============================================================================
  // Events
  // ============================================================================
  events: {
    title: 'Eventos',
    description: 'Consolida la demanda de un evento y compara insumos contra el inventario',
    storedTitle: 'Evento guardado',
    storedDescription: 'Calcula las líneas de demanda persistidas, por ejemplo EVT001',
    inlineTitle: 'Demanda manual',
    inlineDescription: 'Añade productos y cantidades. Los insumos compartidos se consolidan una sola vez',
    eventId: 'Identificador de evento',
    addDemand: 'Añadir línea',
    removeDemand: 'Quitar línea',
    emptyDemands: 'Añade al menos una línea de demanda con producto y cantidad',
    invalidQuantity: 'Cada línea necesita una cantidad mayor que cero',
    consolidated: 'Insumos consolidados',
    consolidatedCaption: 'Materias primas sumadas en todo el evento, sin doble conteo',
    perLine: 'Desglose por línea',
    perLineCaption: 'Expansión de la línea de demanda',
    shortage: 'Faltante',
    enough: 'Cubierto',
    compareCaption: 'Comparación de insumos consolidados contra existencias',
    need: 'Necesario',
    have: 'En inventario',
    shortBy: 'Faltan {qty} {unit}',
    covered: 'Alcanza: hay {have} y se necesitan {need} {unit}',
    noLines: 'Sin líneas de resultado',
    modeLegend: 'Modo de planificación',
    modeStored: 'Usar evento guardado',
    modeInline: 'Capturar demanda',
  },

  // ============================================================================
  // Production
  // ============================================================================
  production: {
    title: 'Producción',
    description: 'Simula el consumo y confirma el descuento de inventario',
    product: 'Producto',
    productPlaceholder: 'Selecciona un producto',
    quantity: 'Cantidad a producir',
    quantityHint: 'Usa un número mayor que cero. La cantidad se envía como texto decimal.',
    simulate: 'Simular consumo',
    simulating: 'Simulando…',
    previewBadge: 'Vista previa — no se ha descontado inventario',
    previewHint: 'Esta simulación es de solo lectura. El inventario real no cambia hasta que confirmes.',
    committedBadge: 'Producción confirmada — inventario actualizado',
    consumed: 'Materiales a consumir',
    leftovers: 'Sobrante proyectado',
    leftoversCommitted: 'Sobrante después del descuento',
    continueConfirm: 'Continuar a confirmación',
    dialogTitle: 'Confirmar producción',
    dialogBody: 'Esta acción escribe en el inventario. La simulación anterior no descontó existencias.',
    confirmWrite: 'Confirmar y descontar',
    confirming: 'Confirmando…',
    comparison: 'Inventario antes y después',
    comparisonPreview: 'Comparación proyectada (aún no escrita)',
    comparisonCommitted: 'Comparación real después de confirmar',
    colProduct: 'Producto',
    colBefore: 'Antes',
    colConsumed: 'Consumo',
    colAfter: 'Después',
    deducted: 'Descontado',
    historyTitle: 'Historial de movimientos',
    historyEmpty: 'Aún no hay movimientos de confirmación.',
    colChange: 'Cambio',
    colBalance: 'Saldo',
    colReason: 'Motivo',
    colWhen: 'Fecha',
    reasonConfirm: 'Confirmación de producción',
    insufficientInventory: 'Inventario insuficiente. La simulación no descontó existencias.',
    confirmConflict: 'Esta confirmación ya está en curso o la clave no coincide. Reintenta la misma solicitud.',
    retryHint: 'El reintento usa la misma clave de idempotencia para no duplicar el descuento.',
    dismiss: 'Cerrar aviso',
  },
} as const;

/** Type-safe translation key accessor */
export type TranslationKey = typeof es;
