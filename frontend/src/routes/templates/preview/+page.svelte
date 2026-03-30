<script lang="ts">
	/**
	 * Template Components Preview Page
	 * For testing and styling the app template component system
	 */

	import '$lib/templates/tokens/tokens.css';

	import {
		// Primitives
		TemplateButton,
		TemplateBadge,
		TemplateInput,
		TemplateSelect,
		TemplateModal,
		TemplateTooltip,
		// Cells
		TextCell,
		NumberCell,
		CurrencyCell,
		DateCell,
		StatusBadge,
		EmailCell,
		PhoneCell,
		URLCell,
		CheckboxCell,
		RatingCell,
		ProgressCell,
		UserCell,
		MultiSelectCell,
		// Views
		DataTable,
		// Chrome
		AppShell,
		AppRenderer,
		// Types
		type AppConfig,
		type Field,
		type RecordData
	} from '$lib/templates';

	// Sample data for testing
	const sampleUsers = [
		{ id: '1', name: 'John Smith', email: 'john@acme.com', avatar: undefined },
		{ id: '2', name: 'Sarah Chen', email: 'sarah@acme.com', avatar: undefined },
		{ id: '3', name: 'Mike Johnson', email: 'mike@acme.com', avatar: undefined }
	];

	const statusOptions = [
		{ value: 'lead', label: 'Lead', color: '#6b7280' },
		{ value: 'qualified', label: 'Qualified', color: '#3b82f6' },
		{ value: 'proposal', label: 'Proposal', color: '#f59e0b' },
		{ value: 'negotiation', label: 'Negotiation', color: '#8b5cf6' },
		{ value: 'won', label: 'Won', color: '#10b981' },
		{ value: 'lost', label: 'Lost', color: '#ef4444' }
	];

	const tagOptions = [
		{ value: 'enterprise', label: 'Enterprise', color: '#8b5cf6' },
		{ value: 'startup', label: 'Startup', color: '#10b981' },
		{ value: 'smb', label: 'SMB', color: '#3b82f6' },
		{ value: 'priority', label: 'Priority', color: '#ef4444' }
	];

	// Sample fields for DataTable
	const sampleFields: Field[] = [
		{ id: 'name', name: 'Company', type: 'text', width: 200 },
		{ id: 'contact', name: 'Contact', type: 'user', width: 150 },
		{ id: 'email', name: 'Email', type: 'email', width: 200 },
		{ id: 'phone', name: 'Phone', type: 'phone', width: 140 },
		{ id: 'status', name: 'Status', type: 'status', options: statusOptions, width: 120 },
		{ id: 'value', name: 'Deal Value', type: 'currency', currency: 'USD', width: 120 },
		{ id: 'probability', name: 'Probability', type: 'progress', width: 140 },
		{ id: 'rating', name: 'Priority', type: 'rating', max: 5, width: 120 },
		{ id: 'tags', name: 'Tags', type: 'multiselect', options: tagOptions, width: 180 },
		{ id: 'website', name: 'Website', type: 'url', width: 200 },
		{ id: 'lastContact', name: 'Last Contact', type: 'date', width: 140 },
		{ id: 'active', name: 'Active', type: 'checkbox', width: 80 }
	];

	// Sample data
	const sampleData: RecordData[] = [
		{
			id: '1',
			name: 'Acme Corporation',
			contact: sampleUsers[0],
			email: 'contact@acme.com',
			phone: '5551234567',
			status: 'proposal',
			value: 125000,
			probability: 75,
			rating: 5,
			tags: ['enterprise', 'priority'],
			website: 'https://acme.com',
			lastContact: '2026-01-28',
			active: true
		},
		{
			id: '2',
			name: 'Beta Industries',
			contact: sampleUsers[1],
			email: 'sales@beta.io',
			phone: '5559876543',
			status: 'qualified',
			value: 45000,
			probability: 40,
			rating: 3,
			tags: ['startup'],
			website: 'https://beta.io',
			lastContact: '2026-01-25',
			active: true
		},
		{
			id: '3',
			name: 'Gamma Technologies',
			contact: sampleUsers[2],
			email: 'info@gamma.tech',
			phone: '5555551212',
			status: 'negotiation',
			value: 89000,
			probability: 85,
			rating: 4,
			tags: ['smb', 'priority'],
			website: 'https://gamma.tech',
			lastContact: '2026-01-30',
			active: true
		},
		{
			id: '4',
			name: 'Delta Services',
			contact: [sampleUsers[0], sampleUsers[1]],
			email: 'hello@delta.co',
			phone: '5558889999',
			status: 'won',
			value: 200000,
			probability: 100,
			rating: 5,
			tags: ['enterprise'],
			website: 'https://delta.co',
			lastContact: '2026-01-20',
			active: true
		},
		{
			id: '5',
			name: 'Epsilon Labs',
			contact: null,
			email: 'team@epsilon.dev',
			phone: '',
			status: 'lead',
			value: 15000,
			probability: 15,
			rating: 2,
			tags: ['startup'],
			website: '',
			lastContact: '2026-01-15',
			active: false
		}
	];

	// Full app config for AppRenderer demo
	const appConfig: AppConfig = {
		id: 'crm-demo',
		version: '1.0.0',
		branding: {
			name: 'Sales Pipeline',
			description: 'Track your deals',
			icon: '📈'
		},
		toolbar: {
			showSearch: true,
			searchPlaceholder: 'Search deals...',
			showViewSwitcher: true,
			showFilter: true,
			showSort: true,
			showExport: true
		},
		fields: sampleFields,
		views: [
			{
				id: 'table',
				name: 'Table',
				type: 'table',
				density: 'comfortable',
				showCheckboxes: true,
				showRowNumbers: true,
				enableInlineEdit: true
			},
			{
				id: 'card',
				name: 'Cards',
				type: 'card',
				titleField: 'name',
				subtitleField: 'email'
			},
			{
				id: 'kanban',
				name: 'Kanban',
				type: 'kanban',
				groupByField: 'status',
				titleField: 'name'
			}
		],
		defaultViewId: 'table',
		features: {
			enableSearch: true,
			enableFilter: true,
			enableSort: true,
			enableInlineEdit: true,
			enableBulkActions: true
		}
	};

	// State for interactive demos
	let textValue = $state('Editable text');
	let numberValue = $state(1234.56);
	let currencyValue = $state(99999.99);
	let dateValue = $state('2026-01-30');
	let statusValue = $state('proposal');
	let checkboxValue = $state(true);
	let ratingValue = $state(4);
	let multiSelectValue = $state(['enterprise', 'priority']);
	let modalOpen = $state(false);

	let selectedSection = $state('all');

	const sections = [
		{ id: 'all', label: 'All Components' },
		{ id: 'primitives', label: 'Primitives' },
		{ id: 'cells', label: 'Cells' },
		{ id: 'table', label: 'DataTable' },
		{ id: 'app', label: 'Full App' }
	];
</script>

<svelte:head>
	<title>Template Components Preview | BusinessOS</title>
</svelte:head>

<div class="preview-page">
	<header class="preview-header">
		<h1>App Template Components</h1>
		<p>Preview and test the template component system</p>
		<nav class="preview-nav">
			{#each sections as section}
				<button
					class="preview-nav-btn"
					class:active={selectedSection === section.id}
					onclick={() => (selectedSection = section.id)}
				>
					{section.label}
				</button>
			{/each}
		</nav>
	</header>

	<main class="preview-content">
		<!-- PRIMITIVES -->
		{#if selectedSection === 'all' || selectedSection === 'primitives'}
			<section class="preview-section">
				<h2>Primitives</h2>

				<div class="preview-group">
					<h3>Button Variants</h3>
					<div class="preview-row">
						<TemplateButton variant="primary">Primary</TemplateButton>
						<TemplateButton variant="secondary">Secondary</TemplateButton>
						<TemplateButton variant="outline">Outline</TemplateButton>
						<TemplateButton variant="ghost">Ghost</TemplateButton>
						<TemplateButton variant="danger">Danger</TemplateButton>
						<TemplateButton variant="success">Success</TemplateButton>
					</div>
					<h3>Button Sizes</h3>
					<div class="preview-row">
						<TemplateButton size="xs">Extra Small</TemplateButton>
						<TemplateButton size="sm">Small</TemplateButton>
						<TemplateButton size="md">Medium</TemplateButton>
						<TemplateButton size="lg">Large</TemplateButton>
					</div>
					<h3>Button States</h3>
					<div class="preview-row">
						<TemplateButton loading={true}>Loading</TemplateButton>
						<TemplateButton disabled={true}>Disabled</TemplateButton>
						<TemplateButton fullWidth={true}>Full Width</TemplateButton>
					</div>
					<h3>Icon-Only Buttons</h3>
					<div class="preview-row">
						<TemplateButton variant="secondary" size="xs" iconOnly={true}>
							<svg viewBox="0 0 20 20" fill="currentColor"><path d="M10 5a1 1 0 011 1v3h3a1 1 0 110 2h-3v3a1 1 0 11-2 0v-3H6a1 1 0 110-2h3V6a1 1 0 011-1z" /></svg>
						</TemplateButton>
						<TemplateButton variant="secondary" size="sm" iconOnly={true}>
							<svg viewBox="0 0 20 20" fill="currentColor"><path d="M10 5a1 1 0 011 1v3h3a1 1 0 110 2h-3v3a1 1 0 11-2 0v-3H6a1 1 0 110-2h3V6a1 1 0 011-1z" /></svg>
						</TemplateButton>
						<TemplateButton variant="secondary" size="md" iconOnly={true}>
							<svg viewBox="0 0 20 20" fill="currentColor"><path d="M10 5a1 1 0 011 1v3h3a1 1 0 110 2h-3v3a1 1 0 11-2 0v-3H6a1 1 0 110-2h3V6a1 1 0 011-1z" /></svg>
						</TemplateButton>
						<TemplateButton variant="secondary" size="lg" iconOnly={true}>
							<svg viewBox="0 0 20 20" fill="currentColor"><path d="M10 5a1 1 0 011 1v3h3a1 1 0 110 2h-3v3a1 1 0 11-2 0v-3H6a1 1 0 110-2h3V6a1 1 0 011-1z" /></svg>
						</TemplateButton>
					</div>
				</div>

				<div class="preview-group">
					<h3>Badge Variants</h3>
					<div class="preview-row">
						<TemplateBadge>Default</TemplateBadge>
						<TemplateBadge variant="success">Success</TemplateBadge>
						<TemplateBadge variant="warning">Warning</TemplateBadge>
						<TemplateBadge variant="error">Error</TemplateBadge>
						<TemplateBadge variant="info">Info</TemplateBadge>
						<TemplateBadge variant="outline">Outline</TemplateBadge>
					</div>
					<h3>Badge Sizes</h3>
					<div class="preview-row">
						<TemplateBadge size="xs">Extra Small</TemplateBadge>
						<TemplateBadge size="sm">Small</TemplateBadge>
						<TemplateBadge size="md">Medium</TemplateBadge>
					</div>
					<h3>Badge Features</h3>
					<div class="preview-row">
						<TemplateBadge dot={true}>With Dot</TemplateBadge>
						<TemplateBadge dot={true} variant="success">Success Dot</TemplateBadge>
						<TemplateBadge removable={true}>Removable</TemplateBadge>
					</div>
					<h3>Custom Badges</h3>
					<div class="preview-row">
						<TemplateBadge variant="custom" backgroundColor="#dbeafe" color="#1e40af">Custom Blue</TemplateBadge>
						<TemplateBadge variant="custom" backgroundColor="#fef3c7" color="#92400e">Custom Orange</TemplateBadge>
						<TemplateBadge variant="custom" backgroundColor="#f3e8ff" color="#7c3aed">Custom Purple</TemplateBadge>
					</div>
				</div>

				<div class="preview-group">
					<h3>Input Sizes</h3>
					<div class="preview-row" style="flex-direction: column; align-items: stretch; max-width: 300px; gap: 8px;">
						<TemplateInput size="sm" placeholder="Small input" />
						<TemplateInput size="md" placeholder="Medium input (default)" />
						<TemplateInput size="lg" placeholder="Large input" />
					</div>
					<h3>Input Features</h3>
					<div class="preview-row" style="flex-direction: column; align-items: stretch; max-width: 300px; gap: 8px;">
						<TemplateInput placeholder="With prefix" prefix="$" />
						<TemplateInput placeholder="With suffix" suffix=".00" />
						<TemplateInput placeholder="Both" prefix="$" suffix="USD" />
						<TemplateInput placeholder="With error" error="This field is required" />
						<TemplateInput placeholder="Disabled" disabled={true} />
					</div>
				</div>

				<div class="preview-group">
					<h3>Select</h3>
					<div class="preview-row" style="max-width: 300px;">
						<TemplateSelect
							options={statusOptions}
							placeholder="Select status..."
						/>
					</div>
				</div>

				<div class="preview-group">
					<h3>Tooltip</h3>
					<div class="preview-row">
						<TemplateTooltip text="Tooltip on top" position="top">
							<TemplateButton variant="secondary">Top</TemplateButton>
						</TemplateTooltip>
						<TemplateTooltip text="Tooltip on bottom" position="bottom">
							<TemplateButton variant="secondary">Bottom</TemplateButton>
						</TemplateTooltip>
						<TemplateTooltip text="Tooltip on left" position="left">
							<TemplateButton variant="secondary">Left</TemplateButton>
						</TemplateTooltip>
						<TemplateTooltip text="Tooltip on right" position="right">
							<TemplateButton variant="secondary">Right</TemplateButton>
						</TemplateTooltip>
					</div>
				</div>

				<div class="preview-group">
					<h3>Modal</h3>
					<div class="preview-row">
						<TemplateButton onclick={() => (modalOpen = true)}>Open Modal</TemplateButton>
					</div>
					<TemplateModal bind:open={modalOpen} title="Sample Modal">
						<p>This is a sample modal dialog. You can put any content here.</p>
						<div style="margin-top: 16px; display: flex; gap: 8px; justify-content: flex-end;">
							<TemplateButton variant="secondary" onclick={() => (modalOpen = false)}>Cancel</TemplateButton>
							<TemplateButton variant="primary" onclick={() => (modalOpen = false)}>Confirm</TemplateButton>
						</div>
					</TemplateModal>
				</div>
			</section>
		{/if}

		<!-- CELLS -->
		{#if selectedSection === 'all' || selectedSection === 'cells'}
			<section class="preview-section">
				<h2>Cell Components</h2>

				<div class="preview-group">
					<h3>Text & Numbers</h3>
					<div class="preview-cells">
						<div class="preview-cell-row">
							<span class="preview-cell-label">TextCell:</span>
							<TextCell value={textValue} editable={true} onchange={(v) => (textValue = v)} />
						</div>
						<div class="preview-cell-row">
							<span class="preview-cell-label">NumberCell:</span>
							<NumberCell value={numberValue} editable={true} precision={2} onchange={(v) => (numberValue = v)} />
						</div>
						<div class="preview-cell-row">
							<span class="preview-cell-label">CurrencyCell:</span>
							<CurrencyCell value={currencyValue} editable={true} onchange={(v) => (currencyValue = v)} />
						</div>
					</div>
				</div>

				<div class="preview-group">
					<h3>Date & Links</h3>
					<div class="preview-cells">
						<div class="preview-cell-row">
							<span class="preview-cell-label">DateCell:</span>
							<DateCell value={dateValue} editable={true} onchange={(v) => (dateValue = v)} />
						</div>
						<div class="preview-cell-row">
							<span class="preview-cell-label">EmailCell:</span>
							<EmailCell value="hello@example.com" />
						</div>
						<div class="preview-cell-row">
							<span class="preview-cell-label">PhoneCell:</span>
							<PhoneCell value="5551234567" />
						</div>
						<div class="preview-cell-row">
							<span class="preview-cell-label">URLCell:</span>
							<URLCell value="https://github.com/anthropics/claude-code" />
						</div>
					</div>
				</div>

				<div class="preview-group">
					<h3>Status & Selection</h3>
					<div class="preview-cells">
						<div class="preview-cell-row">
							<span class="preview-cell-label">StatusBadge:</span>
							<StatusBadge
								value={statusValue}
								options={statusOptions}
								editable={true}
								onchange={(v) => (statusValue = v)}
							/>
						</div>
						<div class="preview-cell-row">
							<span class="preview-cell-label">CheckboxCell:</span>
							<CheckboxCell value={checkboxValue} editable={true} onchange={(v) => (checkboxValue = v)} />
						</div>
						<div class="preview-cell-row">
							<span class="preview-cell-label">MultiSelectCell:</span>
							<MultiSelectCell
								value={multiSelectValue}
								options={tagOptions}
								editable={true}
								onchange={(v) => (multiSelectValue = v)}
							/>
						</div>
					</div>
				</div>

				<div class="preview-group">
					<h3>Visual Indicators</h3>
					<div class="preview-cells">
						<div class="preview-cell-row">
							<span class="preview-cell-label">RatingCell:</span>
							<RatingCell value={ratingValue} editable={true} onchange={(v) => (ratingValue = v)} />
						</div>
						<div class="preview-cell-row">
							<span class="preview-cell-label">ProgressCell:</span>
							<ProgressCell value={75} />
						</div>
						<div class="preview-cell-row">
							<span class="preview-cell-label">UserCell (single):</span>
							<UserCell value={sampleUsers[0]} showEmail={true} />
						</div>
						<div class="preview-cell-row">
							<span class="preview-cell-label">UserCell (multi):</span>
							<UserCell value={sampleUsers} />
						</div>
					</div>
				</div>
			</section>
		{/if}

		<!-- DATATABLE -->
		{#if selectedSection === 'all' || selectedSection === 'table'}
			<section class="preview-section">
				<h2>DataTable View</h2>
				<div class="preview-table-container">
					<DataTable
						fields={sampleFields}
						data={sampleData}
						config={{
							density: 'comfortable',
							showCheckboxes: true,
							showRowNumbers: true,
							enableInlineEdit: true
						}}
					/>
				</div>
			</section>
		{/if}

		<!-- FULL APP -->
		{#if selectedSection === 'all' || selectedSection === 'app'}
			<section class="preview-section">
				<h2>Full App Demo</h2>
				<p class="preview-description">
					This is a complete app rendered from a JSON config. It includes the shell, toolbar, and data table.
				</p>
				<div class="preview-app-container">
					<AppRenderer
						config={appConfig}
						data={sampleData}
						onrecordclick={(record) => { if (import.meta.env.DEV) console.log('Clicked:', record); }}
						onrecordcreate={() => { if (import.meta.env.DEV) console.log('Create new record'); }}
						onrecordedit={(id, field, value) => { if (import.meta.env.DEV) console.log('Edit:', id, field, value); }}
					/>
				</div>
			</section>
		{/if}
	</main>
</div>

<style>
	.preview-page {
		min-height: 100vh;
		background: #f8fafc;
		font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
	}

	.preview-header {
		padding: 32px 48px;
		background: white;
		border-bottom: 1px solid #e2e8f0;
	}

	.preview-header h1 {
		margin: 0 0 8px;
		font-size: 28px;
		font-weight: 700;
		color: #0f172a;
	}

	.preview-header p {
		margin: 0 0 24px;
		color: #64748b;
	}

	.preview-nav {
		display: flex;
		gap: 8px;
	}

	.preview-nav-btn {
		padding: 8px 16px;
		background: #f1f5f9;
		border: none;
		border-radius: 8px;
		font-size: 14px;
		font-weight: 500;
		color: #475569;
		cursor: pointer;
		transition: all 0.15s;
	}

	.preview-nav-btn:hover {
		background: #e2e8f0;
	}

	.preview-nav-btn.active {
		background: #0f172a;
		color: white;
	}

	.preview-content {
		padding: 32px 48px;
	}

	.preview-section {
		margin-bottom: 48px;
	}

	.preview-section h2 {
		margin: 0 0 24px;
		font-size: 20px;
		font-weight: 600;
		color: #0f172a;
		padding-bottom: 12px;
		border-bottom: 2px solid #e2e8f0;
	}

	.preview-group {
		margin-bottom: 32px;
	}

	.preview-group h3 {
		margin: 0 0 12px;
		font-size: 14px;
		font-weight: 600;
		color: #64748b;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.preview-row {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 12px;
	}

	.preview-cells {
		display: flex;
		flex-direction: column;
		gap: 8px;
		background: white;
		border: 1px solid #e2e8f0;
		border-radius: 12px;
		padding: 16px;
	}

	.preview-cell-row {
		display: flex;
		align-items: center;
		gap: 16px;
	}

	.preview-cell-label {
		width: 140px;
		font-size: 13px;
		font-weight: 500;
		color: #64748b;
	}

	.preview-table-container {
		background: white;
		border-radius: 12px;
		overflow: hidden;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
	}

	.preview-description {
		margin: 0 0 16px;
		color: #64748b;
		font-size: 14px;
	}

	.preview-app-container {
		height: 600px;
		border-radius: 12px;
		overflow: hidden;
		box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
	}
</style>
