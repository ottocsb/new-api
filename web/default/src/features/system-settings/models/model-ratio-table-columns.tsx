import { type ColumnDef } from '@tanstack/react-table'
import { Pencil, Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { DataTableColumnHeader } from '@/components/data-table'
import { StatusBadge } from '@/components/status-badge'
import {
  getModeLabel,
  getModeVariant,
  getPriceDetail,
  getPriceSummary,
  type ModelRow,
} from './model-pricing-snapshots'

const filterBySelectedValues = (
  rowValue: unknown,
  filterValue: unknown
): boolean => {
  if (!Array.isArray(filterValue) || filterValue.length === 0) return true
  return filterValue.includes(String(rowValue))
}

type BuildModelRatioColumnsOptions = {
  onDelete: (name: string) => void
  onEdit: (model: ModelRow) => void
  t: (key: string) => string
}

export function buildModelRatioColumns({
  onDelete,
  onEdit,
  t,
}: BuildModelRatioColumnsOptions): ColumnDef<ModelRow>[] {
  return [
    {
      id: 'select',
      header: ({ table }) => (
        <Checkbox
          checked={table.getIsAllPageRowsSelected()}
          indeterminate={table.getIsSomePageRowsSelected()}
          onCheckedChange={(value) => table.toggleAllPageRowsSelected(!!value)}
          aria-label={t('Select all')}
          className='translate-y-[2px]'
        />
      ),
      cell: ({ row }) => (
        <Checkbox
          checked={row.getIsSelected()}
          onCheckedChange={(value) => row.toggleSelected(!!value)}
          aria-label={t('Select row')}
          className='translate-y-[2px]'
        />
      ),
      enableSorting: false,
      enableHiding: false,
      meta: { label: t('Select') },
    },
    {
      accessorKey: 'name',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title={t('Model name')} />
      ),
      cell: ({ row }) => (
        <div className='flex items-center gap-2 font-medium'>
          {row.getValue('name')}
          {row.original.billingMode === 'tiered_expr' && (
            <StatusBadge label={t('Tiered')} variant='info' copyable={false} />
          )}
          {row.original.hasConflict && (
            <StatusBadge
              label={t('Conflict')}
              variant='danger'
              copyable={false}
            />
          )}
        </div>
      ),
      enableHiding: false,
    },
    {
      accessorKey: 'billingMode',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title={t('Mode')} />
      ),
      cell: ({ row }) => (
        <StatusBadge
          label={t(getModeLabel(row.original.billingMode))}
          variant={getModeVariant(row.original.billingMode)}
          copyable={false}
          showDot={false}
          className='px-0'
        />
      ),
      filterFn: (row, id, value) =>
        filterBySelectedValues(row.getValue(id), value),
      meta: { label: t('Mode') },
    },
    {
      id: 'priceSummary',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title={t('Price summary')} />
      ),
      cell: ({ row }) => (
        <div className='flex min-w-[180px] flex-col gap-1'>
          <span className='font-medium'>
            {getPriceSummary(row.original, t)}
          </span>
          <span className='text-muted-foreground max-w-[320px] truncate text-xs'>
            {getPriceDetail(row.original, t)}
          </span>
        </div>
      ),
      sortingFn: (rowA, rowB) =>
        getPriceSummary(rowA.original, t).localeCompare(
          getPriceSummary(rowB.original, t)
        ),
      meta: { label: t('Price summary') },
    },
    {
      id: 'actions',
      header: () => <div className='text-right'>{t('Actions')}</div>,
      cell: ({ row }) => (
        <div className='flex justify-end gap-2'>
          <Button
            variant='ghost'
            size='sm'
            onClick={() => onEdit(row.original)}
          >
            <Pencil />
          </Button>
          <Button
            variant='ghost'
            size='sm'
            onClick={() => onDelete(row.original.name)}
          >
            <Trash2 />
          </Button>
        </div>
      ),
      enableHiding: false,
    },
  ]
}
