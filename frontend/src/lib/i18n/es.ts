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
    placeholder: 'El panel operativo se completará en commit 8',
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
  },

  // ============================================================================
  // Product types
  // ============================================================================
  productType: {
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
  },

  // ============================================================================
  // Forms
  // ============================================================================
  forms: {
    required: 'Este campo es obligatorio',
    invalidFormat: 'Formato inválido',
    mustBePositive: 'Debe ser un número positivo',
    mustBeGreaterThanZero: 'Debe ser mayor que cero',
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
  },
} as const;

/** Type-safe translation key accessor */
export type TranslationKey = typeof es;
