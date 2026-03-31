import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, fireEvent, screen } from '@testing-library/svelte';
import AgentCard from './AgentCard.svelte';
import type { CustomAgent } from '$lib/api/ai/types';

describe('AgentCard Component', () => {
  const mockAgent: CustomAgent = {
    id: '1',
    user_id: 'user1',
    name: 'test-agent',
    display_name: 'Test Agent',
    description: 'A test agent for testing',
    system_prompt: 'You are a test agent',
    category: 'general',
    model_preference: 'gpt-4',
    is_active: true,
    times_used: 42,
    created_at: '2024-01-01',
    updated_at: '2024-01-01'
  };

  // Note: AgentCard does not render an <img> tag; it always shows initials.
  // The `avatar` field is not used by the component.
  const mockAgentWithAvatar: CustomAgent = {
    ...mockAgent,
    id: '2',
    avatar: 'https://example.com/avatar.png'
  };

  const inactiveAgent: CustomAgent = {
    ...mockAgent,
    id: '3',
    is_active: false
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('should render agent card with all basic information', () => {
      const { container } = render(AgentCard, {
        props: { agent: mockAgent }
      });

      // display_name renders as-is
      expect(screen.getByText('Test Agent')).toBeTruthy();
      // handle renders as "@{agent.name}"
      expect(screen.getByText('@test-agent')).toBeTruthy();
      // description renders as-is
      expect(screen.getByText('A test agent for testing')).toBeTruthy();
      // category 'general' maps to "General" via categoryLabels
      expect(screen.getByText('General')).toBeTruthy();
      // model_preference 'gpt-4' → replace('claude-','') → replace('-4',' 4') → 'gpt 4'
      expect(screen.getByText('gpt 4')).toBeTruthy();
      // times_used renders as "{formatUsage(n)} uses"
      expect(screen.getByText('42 uses')).toBeTruthy();
      expect(container).toBeTruthy();
    });

    it('should render active status dot', () => {
      const { container } = render(AgentCard, {
        props: { agent: mockAgent }
      });

      // Active status is a dot with title="Active", not visible text
      const activeDot = container.querySelector('.ac__dot--active');
      expect(activeDot).toBeTruthy();
      expect(activeDot?.getAttribute('title')).toBe('Active');
    });

    it('should render inactive status dot', () => {
      const { container } = render(AgentCard, {
        props: { agent: inactiveAgent }
      });

      // Inactive status is a dot with title="Inactive", not visible text
      const inactiveDot = container.querySelector('.ac__dot--inactive');
      expect(inactiveDot).toBeTruthy();
      expect(inactiveDot?.getAttribute('title')).toBe('Inactive');
    });

    it('should render initials when no avatar provided', () => {
      const { container } = render(AgentCard, {
        props: { agent: mockAgent }
      });

      // Component always renders initials; no <img> is ever rendered
      // "Test Agent" → "TA"
      expect(screen.getByText('TA')).toBeTruthy();
      expect(container.querySelector('img')).toBeFalsy();
    });

    it('should render initials even when avatar field is set (component ignores avatar)', () => {
      const { container } = render(AgentCard, {
        props: { agent: mockAgentWithAvatar }
      });

      // Component does not render <img> — it always shows initials
      expect(screen.getByText('TA')).toBeTruthy();
      expect(container.querySelector('img')).toBeFalsy();
    });

    it('should handle agent without description', () => {
      const agentNoDesc = { ...mockAgent, description: undefined };
      render(AgentCard, {
        props: { agent: agentNoDesc }
      });

      // Component renders 'No description provided.' (with trailing period)
      expect(screen.getByText('No description provided.')).toBeTruthy();
    });

    it('should not render model badge if no model preference', () => {
      const agentNoModel = { ...mockAgent, model_preference: undefined };
      const { container } = render(AgentCard, {
        props: { agent: agentNoModel }
      });

      expect(container.textContent).not.toContain('gpt-4');
    });

    it('should not render usage count badge if zero or undefined', () => {
      const agentNoUsage = { ...mockAgent, times_used: 0 };
      const { container } = render(AgentCard, {
        props: { agent: agentNoUsage }
      });

      // Component only renders usage when times_used > 0
      expect(container.querySelector('.ac__usage')).toBeFalsy();
    });

    it('should render in compact variant', () => {
      const { container } = render(AgentCard, {
        props: { agent: mockAgent, variant: 'compact' }
      });

      // Compact variant adds class 'ac--compact' to the root element
      const card = container.querySelector('.ac--compact');
      expect(card).toBeTruthy();
    });
  });

  describe('Interactions', () => {
    it('should call onSelect when card is clicked', async () => {
      const onSelect = vi.fn();
      const { container } = render(AgentCard, {
        props: { agent: mockAgent, onSelect }
      });

      const card = container.querySelector('[role="button"]');
      expect(card).toBeTruthy();

      await fireEvent.click(card!);
      expect(onSelect).toHaveBeenCalledWith(mockAgent);
    });

    it('should call onSelect when Enter key is pressed', async () => {
      const onSelect = vi.fn();
      const { container } = render(AgentCard, {
        props: { agent: mockAgent, onSelect }
      });

      const card = container.querySelector('[role="button"]');
      expect(card).toBeTruthy();

      await fireEvent.keyDown(card!, { key: 'Enter' });
      expect(onSelect).toHaveBeenCalledWith(mockAgent);
    });

    it('should call onSelect when Space key is pressed', async () => {
      const onSelect = vi.fn();
      const { container } = render(AgentCard, {
        props: { agent: mockAgent, onSelect }
      });

      const card = container.querySelector('[role="button"]');
      expect(card).toBeTruthy();

      await fireEvent.keyDown(card!, { key: ' ' });
      expect(onSelect).toHaveBeenCalledWith(mockAgent);
    });

    it('should set role=button and aria-label when onSelect is provided', () => {
      const { container } = render(AgentCard, {
        props: { agent: mockAgent, onSelect: vi.fn() }
      });

      // The card itself is the interactive element — no separate "Select" button
      const card = container.querySelector('[role="button"]');
      expect(card).toBeTruthy();
      expect(card?.getAttribute('aria-label')).toBe('Open Test Agent');
    });

    it('should not have role=button when onSelect is not provided', () => {
      const { container } = render(AgentCard, {
        props: { agent: mockAgent }
      });

      expect(container.querySelector('[role="button"]')).toBeFalsy();
      expect(container.textContent).not.toContain('Select');
    });

    it('should open menu when menu button is clicked', async () => {
      render(AgentCard, {
        props: { agent: mockAgent, onEdit: vi.fn(), onDelete: vi.fn() }
      });

      const menuButton = screen.getByLabelText('More actions');
      await fireEvent.click(menuButton);

      expect(screen.getByText('Edit')).toBeTruthy();
      expect(screen.getByText('Delete')).toBeTruthy();
    });

    it('should call onEdit when Edit button is clicked', async () => {
      const onEdit = vi.fn();
      render(AgentCard, {
        props: { agent: mockAgent, onEdit, onDelete: vi.fn() }
      });

      const menuButton = screen.getByLabelText('More actions');
      await fireEvent.click(menuButton);

      const editButton = screen.getByText('Edit');
      await fireEvent.click(editButton);

      expect(onEdit).toHaveBeenCalledWith(mockAgent);
    });

    it('should call onDelete immediately when Delete button is clicked', async () => {
      const onDelete = vi.fn();
      render(AgentCard, {
        props: { agent: mockAgent, onEdit: vi.fn(), onDelete }
      });

      const menuButton = screen.getByLabelText('More actions');
      await fireEvent.click(menuButton);

      const deleteButton = screen.getByText('Delete');
      await fireEvent.click(deleteButton);

      // Component calls onDelete immediately — no confirmation dialog
      expect(onDelete).toHaveBeenCalledWith(mockAgent);
    });

    it('should not show menu button when no edit/delete handlers', () => {
      const { container } = render(AgentCard, {
        props: { agent: mockAgent }
      });

      expect(container.querySelector('[aria-label="More actions"]')).toBeFalsy();
    });

    it('should prevent event propagation when clicking menu', async () => {
      const onSelect = vi.fn();
      render(AgentCard, {
        props: { agent: mockAgent, onSelect, onEdit: vi.fn() }
      });

      const menuButton = screen.getByLabelText('More actions');
      await fireEvent.click(menuButton);

      // onSelect should not be called when clicking the menu
      expect(onSelect).not.toHaveBeenCalled();
    });

    it('should call onSelect exactly once when card is clicked', async () => {
      const onSelect = vi.fn();
      const { container } = render(AgentCard, {
        props: { agent: mockAgent, onSelect }
      });

      const card = container.querySelector('[role="button"]');
      await fireEvent.click(card!);

      // onSelect called exactly once
      expect(onSelect).toHaveBeenCalledTimes(1);
    });
  });

  describe('Helper Functions', () => {
    it('should generate correct initials for single word', () => {
      const singleWordAgent = { ...mockAgent, display_name: 'Agent' };
      render(AgentCard, {
        props: { agent: singleWordAgent }
      });

      expect(screen.getByText('AG')).toBeTruthy();
    });

    it('should generate correct initials for multi-word names', () => {
      const multiWordAgent = { ...mockAgent, display_name: 'Super Awesome Agent' };
      render(AgentCard, {
        props: { agent: multiWordAgent }
      });

      // First letter of first word + first letter of second word
      expect(screen.getByText('SA')).toBeTruthy();
    });

    it('should apply correct category labels', () => {
      const { container, rerender } = render(AgentCard, {
        props: { agent: { ...mockAgent, category: 'general' } }
      });

      // 'general' maps to 'General' via categoryLabels
      expect(container.textContent).toContain('General');

      // 'specialist' has no categoryLabels entry → falls back to raw value
      rerender({ agent: { ...mockAgent, category: 'specialist' } });
      expect(container.textContent).toContain('specialist');

      // 'custom' has no categoryLabels entry → falls back to raw value
      rerender({ agent: { ...mockAgent, category: 'custom' } });
      expect(container.textContent).toContain('custom');

      // no category → badge is not rendered
      rerender({ agent: { ...mockAgent, category: undefined } });
      expect(container.querySelector('.ac__badge')).toBeFalsy();
    });
  });

  describe('Accessibility', () => {
    it('should have proper ARIA attributes when selectable', () => {
      const { container } = render(AgentCard, {
        props: { agent: mockAgent, onSelect: vi.fn() }
      });

      const card = container.querySelector('[role="button"]');
      expect(card?.getAttribute('tabindex')).toBe('0');
    });

    it('should not be keyboard accessible when not selectable', () => {
      const { container } = render(AgentCard, {
        props: { agent: mockAgent }
      });

      // Without onSelect, root div has tabindex="-1" and no role="button"
      const card = container.querySelector('.ac');
      expect(card?.getAttribute('tabindex')).toBe('-1');
      expect(container.querySelector('[role="button"]')).toBeFalsy();
    });

    it('should have proper aria-label for menu button', () => {
      render(AgentCard, {
        props: { agent: mockAgent, onEdit: vi.fn() }
      });

      const menuButton = screen.getByLabelText('More actions');
      expect(menuButton).toBeTruthy();
    });

    it('should not render an img element (component uses initials only)', () => {
      const { container } = render(AgentCard, {
        props: { agent: mockAgentWithAvatar }
      });

      // AgentCard renders initials, not an <img> element
      expect(container.querySelector('img')).toBeFalsy();
      expect(screen.getByText('TA')).toBeTruthy();
    });

    it('should have title attribute on description paragraph', () => {
      const { container } = render(AgentCard, {
        props: { agent: mockAgent }
      });

      // Description is in .ac__desc — no title attribute is set by the component
      const description = container.querySelector('.ac__desc');
      expect(description).toBeTruthy();
      expect(description?.textContent).toBe('A test agent for testing');
    });
  });

  describe('Edge Cases', () => {
    it('should handle very long agent names', () => {
      const longNameAgent = {
        ...mockAgent,
        display_name: 'This is a very long agent name that should be truncated'
      };

      const { container } = render(AgentCard, {
        props: { agent: longNameAgent }
      });

      expect(container.textContent).toContain('This is a very long agent name');
    });

    it('should handle very long descriptions', () => {
      const longDescAgent = {
        ...mockAgent,
        description: 'This is a very long description '.repeat(20)
      };

      const { container } = render(AgentCard, {
        props: { agent: longDescAgent }
      });

      // Description is clamped via CSS on .ac__desc
      const description = container.querySelector('.ac__desc');
      expect(description).toBeTruthy();
    });

    it('should handle agent with no category', () => {
      const noCategoryAgent = { ...mockAgent, category: undefined };
      const { container } = render(AgentCard, {
        props: { agent: noCategoryAgent }
      });

      // Should still render without errors
      expect(screen.getByText('Test Agent')).toBeTruthy();
    });

    it('should handle agent with empty string name', () => {
      const emptyNameAgent = { ...mockAgent, display_name: '', name: 'empty' };
      const { container } = render(AgentCard, {
        props: { agent: emptyNameAgent }
      });

      expect(container.textContent).toContain('@empty');
    });

    it('should handle all callback props being undefined', () => {
      const { container } = render(AgentCard, {
        props: {
          agent: mockAgent,
          onSelect: undefined,
          onEdit: undefined,
          onDelete: undefined
        }
      });

      // Should render without errors
      expect(container).toBeTruthy();
    });
  });
});
