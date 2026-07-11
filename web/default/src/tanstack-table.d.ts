import '@tanstack/react-table'

declare module '@tanstack/react-table' {
  interface ColumnMeta<_TData, _TValue> {
    label?: string
    description?: string
    className?: string
    pinned?: 'left' | 'right'
    contentSized?: boolean
    /**
     * Desktop and card content policy. `full` keeps the complete value visible,
     * `wrap` allows natural line wrapping, and `summary` delegates progressive
     * disclosure to the cell.
     */
    contentMode?: 'full' | 'wrap' | 'summary'
    /**
     * Responsive card placement. Secondary fields stay accessible behind the
     * shared details disclosure; only duplicate/technical fields use hidden.
     */
    cardRole?: 'title' | 'badge' | 'primary' | 'secondary' | 'hidden'
    cardOrder?: number
    cardSpan?: 1 | 2
  }
}
