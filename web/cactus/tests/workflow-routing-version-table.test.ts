import { mount } from '@vue/test-utils'
import { defineComponent, h, ref, watch } from 'vue'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import RoutingVersionTable from '../app/components/dag/routing/RoutingVersionTable.vue'
import type { RoutingVersionRowRecord } from '../app/composables/useWorkflowRouting'

describe('workflow routing version table', () => {
  beforeEach(() => {
    vi.unstubAllGlobals()
    vi.stubGlobal('ref', ref)
    vi.stubGlobal('watch', watch)
    vi.stubGlobal('useI18n', () => ({
      t: translate,
    }))
  })

  it('renders schema version as text and highlights only direct native process versions', () => {
    const wrapper = mount(RoutingVersionTable, {
      props: {
        rows: [routingRow(), routingRowWithVersionCode()],
      },
      global: {
        stubs: {
          DataTablePagination: true,
          DropdownMenu: passthrough(),
          DropdownMenuContent: passthrough(),
          DropdownMenuItem: passthrough('button'),
          DropdownMenuTrigger: passthrough(),
        },
      },
    })

    const text = wrapper.text()
    const html = wrapper.html()

    expect(text).toContain('public v5')
    expect(text).not.toContain('v6 v6')
    expect(text).not.toContain('Schema version 5')
    expect(text).toContain('Workflow v1')
    expect(text).not.toContain('Workflow v1 - native')
    expect(text).toContain('Mapper #77')
    expect(html).toContain('bg-zinc-900')

    const schemaInline = wrapper.get('[data-testid="routing-schema-inline-11"]')
    expect(schemaInline.classes()).toEqual(expect.arrayContaining([
      'flex',
      'items-center',
      'gap-2',
      'whitespace-nowrap',
    ]))
  })
})

function routingRow(): RoutingVersionRowRecord {
  return {
    input_schema_id: 11,
    workflow_id: 42,
    input_schema_code: 'public',
    input_schema_version_number: 5,
    input_schema_status: 'active',
    is_default: true,
    supported_versions: [{
      compatibility_id: 31,
      workflow_version_id: 101,
      workflow_version_name: 'Workflow',
      workflow_version_number: 1,
      compatibility_type: 'native',
      support_mode: 'native',
      is_default_route: true,
      is_valid: true,
      is_active: true,
    }, {
      compatibility_id: 32,
      workflow_version_id: 102,
      workflow_version_name: 'Workflow',
      workflow_version_number: 2,
      compatibility_type: 'native',
      workflow_input_mapper_id: 77,
      support_mode: 'mapper',
      is_default_route: false,
      is_valid: true,
      is_active: true,
    }],
    active_tests: [],
  }
}

function routingRowWithVersionCode(): RoutingVersionRowRecord {
  return {
    ...routingRow(),
    input_schema_id: 12,
    input_schema_code: 'v6',
    input_schema_version_number: 6,
    supported_versions: [],
  }
}

function passthrough(tag = 'div') {
  return defineComponent({
    setup(_, { attrs, slots }) {
      return () => h(tag, attrs, slots.default?.())
    },
  })
}

function translate(key: string, params?: Record<string, unknown>) {
  const translations: Record<string, string> = {
    'common.actions': 'Actions',
    'common.delete': 'Delete',
    'common.edit': 'Edit',
    'common.page': `Page ${params?.page ?? 1} of ${params?.total ?? 1}`,
    'common.rowsPerPage': 'Rows per page',
    'workflowRouting.activeTesting': 'Active testing',
    'workflowRouting.compatibility': 'Compatibility',
    'workflowRouting.fieldDefaultSchema': 'Default schema',
    'workflowRouting.native': 'native',
    'workflowRouting.noRoutingVersions': 'No routing versions',
    'workflowRouting.schemaVersion': 'Schema version',
    'workflowRouting.schemaVersionNumber': `Schema version ${params?.number}`,
    'workflowRouting.statusActive': 'Active',
    'workflowRouting.statusArchived': 'Archived',
    'workflowRouting.statusDeprecated': 'Deprecated',
    'workflowRouting.statusDraft': 'Draft',
    'workflowRouting.supportedProcessVersions': 'Supported process versions',
  }
  return translations[key] ?? key
}
