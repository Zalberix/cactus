import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import StepRenameDialog from '../app/components/dag/StepRenameDialog.vue'

const DialogStub = {
  props: ['open'],
  emits: ['update:open'],
  template: '<div><slot /></div>',
}

const ButtonStub = {
  props: ['disabled'],
  template: '<button type="button" :disabled="disabled"><slot /></button>',
}

function mountDialog(props: Record<string, unknown> = {}) {
  vi.stubGlobal('useI18n', () => ({ t: (key: string) => key }))
  return mount(StepRenameDialog, {
    props: {
      open: true,
      nodeId: '12',
      currentName: 'SMTP',
      names: [
        { id: '12', name: 'SMTP' },
        { id: '13', name: 'Delay' },
      ],
      ...props,
    },
    global: {
      stubs: {
        Dialog: DialogStub,
        DialogContent: { template: '<section><slot /></section>' },
        DialogHeader: { template: '<header><slot /></header>' },
        DialogTitle: { template: '<h2><slot /></h2>' },
        DialogFooter: { template: '<footer><slot /></footer>' },
        Button: ButtonStub,
        Input: {
          props: ['modelValue'],
          emits: ['update:modelValue'],
          template: '<input :value="modelValue" @input="$emit(\'update:modelValue\', $event.target.value)" />',
        },
      },
    },
  })
}

describe('StepRenameDialog', () => {
  it('emits a trimmed name when the value is unique', async () => {
    const wrapper = mountDialog()

    await wrapper.get('input').setValue('  SMTP send  ')
    await wrapper.get('[data-testid="save-step-name"]').trigger('click')

    expect(wrapper.emitted('save')).toEqual([['12', 'SMTP send']])
  })

  it('disables save and shows feedback for a duplicate name', async () => {
    const wrapper = mountDialog()

    await wrapper.get('input').setValue('Delay')

    expect(wrapper.get('[data-testid="step-name-error"]').text()).toContain('editor.stepNameDuplicate')
    expect(wrapper.get('[data-testid="save-step-name"]').attributes('disabled')).toBeDefined()
    expect(wrapper.emitted('save')).toBeUndefined()
  })
})
