<script lang="ts">
	import { userAppsStore, type CreateUserAppParams } from '$lib/stores/userAppsStore';
	import { desktop3dStore } from '$lib/stores/desktop3dStore';
	import {
		X,
		Plus,
		Search,
		Globe,
		ExternalLink,
		Trash2,
		Star,
		Grid3X3,
		MessageSquare,
		Palette,
		Code,
		FolderOpen,
		Brain,
		Briefcase,
		TrendingUp,
		Music,
		Users,
		Check,
		Layers,
		Power,
		ToggleLeft,
		ToggleRight
	} from 'lucide-svelte';
	import { createEventDispatcher } from 'svelte';
	import { fade, fly, scale } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';

	const dispatch = createEventDispatcher();

	interface Props {
		workspaceId: string;
		onClose?: () => void;
		isPage?: boolean; // When true, renders as page content without modal wrapper
	}

	let { workspaceId, onClose, isPage = false }: Props = $props();

	// Form state
	let name = $state('');
	let url = $state('');
	let category = $state('productivity');
	let description = $state('');
	let isSubmitting = $state(false);
	let searchQuery = $state('');
	let activeTab = $state<'browse' | 'myapps' | 'custom'>('browse');
	let selectedCategory = $state<string>('all');

	// Reactive list of added app URLs (tracks store changes)
	let addedAppUrls = $derived($userAppsStore.apps.map(app => app.url));

	// Categories with icons and colors - focused on business tools
	const categories = [
		{ id: 'all', name: 'All Apps', icon: Grid3X3, color: '#10B981' },
		{ id: 'ai', name: 'AI Tools', icon: Brain, color: '#8B5CF6' },
		{ id: 'business', name: 'Business & CRM', icon: Briefcase, color: '#F97316' },
		{ id: 'productivity', name: 'Productivity', icon: TrendingUp, color: '#10B981' },
		{ id: 'communication', name: 'Communication', icon: MessageSquare, color: '#3B82F6' },
		{ id: 'project-management', name: 'Project Management', icon: Briefcase, color: '#F59E0B' },
		{ id: 'design', name: 'Design', icon: Palette, color: '#EC4899' },
		{ id: 'storage', name: 'Storage', icon: FolderOpen, color: '#14B8A6' },
		{ id: 'media', name: 'Media', icon: Music, color: '#EF4444' },
		{ id: 'social', name: 'Social', icon: Users, color: '#8B5CF6' }
	];

	// Popular web apps with logos - Business focused
	const popularApps = [
		// Business & CRM - CRM
		{
			name: 'HubSpot',
			url: 'https://app.hubspot.com',
			color: '#FF7A59',
			category: 'business',
			subcategory: 'CRM',
			description: 'CRM & marketing automation',
			logo: 'https://www.google.com/s2/favicons?domain=hubspot.com&sz=128',
			featured: true
		},
		{
			name: 'Salesforce',
			url: 'https://login.salesforce.com',
			color: '#00A1E0',
			category: 'business',
			subcategory: 'CRM',
			description: 'Enterprise CRM platform',
			logo: 'https://www.google.com/s2/favicons?domain=salesforce.com&sz=128',
			featured: true
		},
		{
			name: 'Pipedrive',
			url: 'https://app.pipedrive.com',
			color: '#00B67A',
			category: 'business',
			subcategory: 'CRM',
			description: 'Sales CRM pipeline',
			logo: 'https://www.google.com/s2/favicons?domain=pipedrive.com&sz=128',
			featured: true
		},
		{
			name: 'Zoho CRM',
			url: 'https://crm.zoho.com',
			color: '#D32F2F',
			category: 'business',
			subcategory: 'CRM',
			description: 'Complete CRM solution',
			logo: 'https://www.google.com/s2/favicons?domain=zoho.com&sz=128'
		},
		{
			name: 'Freshsales',
			url: 'https://www.freshworks.com/crm/sales/',
			color: '#5B5FC7',
			category: 'business',
			subcategory: 'CRM',
			description: 'Sales CRM by Freshworks',
			logo: 'https://www.google.com/s2/favicons?domain=freshworks.com&sz=128'
		},
		{
			name: 'Close',
			url: 'https://app.close.com',
			color: '#00C27C',
			category: 'business',
			subcategory: 'CRM',
			description: 'Sales engagement CRM',
			logo: 'https://www.google.com/s2/favicons?domain=close.com&sz=128'
		},
		{
			name: 'Copper',
			url: 'https://app.copper.com',
			color: '#F5A623',
			category: 'business',
			subcategory: 'CRM',
			description: 'CRM for Google Workspace',
			logo: 'https://www.google.com/s2/favicons?domain=copper.com&sz=128'
		},
		{
			name: 'Streak',
			url: 'https://www.streak.com',
			color: '#2185D0',
			category: 'business',
			subcategory: 'CRM',
			description: 'CRM inside Gmail',
			logo: 'https://www.google.com/s2/favicons?domain=streak.com&sz=128'
		},
		{
			name: 'Monday Sales',
			url: 'https://auth.monday.com',
			color: '#FF3D57',
			category: 'business',
			subcategory: 'CRM',
			description: 'Sales CRM',
			logo: 'https://www.google.com/s2/favicons?domain=monday.com&sz=128'
		},
		// Business & CRM - Accounting & Finance
		{
			name: 'Stripe',
			url: 'https://dashboard.stripe.com',
			color: '#635BFF',
			category: 'business',
			subcategory: 'Accounting & Finance',
			description: 'Payment processing',
			logo: 'https://www.google.com/s2/favicons?domain=stripe.com&sz=128',
			featured: true
		},
		{
			name: 'QuickBooks',
			url: 'https://quickbooks.intuit.com',
			color: '#2CA01C',
			category: 'business',
			subcategory: 'Accounting & Finance',
			description: 'Accounting & invoicing',
			logo: 'https://www.google.com/s2/favicons?domain=quickbooks.intuit.com&sz=128'
		},
		{
			name: 'Xero',
			url: 'https://go.xero.com',
			color: '#13B5EA',
			category: 'business',
			subcategory: 'Accounting & Finance',
			description: 'Cloud accounting',
			logo: 'https://www.google.com/s2/favicons?domain=xero.com&sz=128'
		},
		{
			name: 'FreshBooks',
			url: 'https://my.freshbooks.com',
			color: '#00ABF0',
			category: 'business',
			subcategory: 'Accounting & Finance',
			description: 'Invoicing & accounting',
			logo: 'https://www.google.com/s2/favicons?domain=freshbooks.com&sz=128'
		},
		{
			name: 'Wave',
			url: 'https://www.waveapps.com',
			color: '#0047BB',
			category: 'business',
			subcategory: 'Accounting & Finance',
			description: 'Free accounting software',
			logo: 'https://www.google.com/s2/favicons?domain=waveapps.com&sz=128'
		},
		// Business & CRM - Customer Support
		{
			name: 'Intercom',
			url: 'https://app.intercom.com',
			color: '#1F8DED',
			category: 'business',
			subcategory: 'Customer Support',
			description: 'Customer messaging',
			logo: 'https://www.google.com/s2/favicons?domain=intercom.com&sz=128'
		},
		{
			name: 'Zendesk',
			url: 'https://www.zendesk.com',
			color: '#03363D',
			category: 'business',
			subcategory: 'Customer Support',
			description: 'Customer service platform',
			logo: 'https://www.google.com/s2/favicons?domain=zendesk.com&sz=128'
		},
		{
			name: 'Freshdesk',
			url: 'https://freshdesk.com',
			color: '#25C16F',
			category: 'business',
			subcategory: 'Customer Support',
			description: 'Customer support software',
			logo: 'https://www.google.com/s2/favicons?domain=freshdesk.com&sz=128'
		},
		{
			name: 'Drift',
			url: 'https://app.drift.com',
			color: '#0176FF',
			category: 'business',
			subcategory: 'Customer Support',
			description: 'Conversational marketing',
			logo: 'https://www.google.com/s2/favicons?domain=drift.com&sz=128'
		},
		// Business & CRM - Scheduling & Documents
		{
			name: 'Calendly',
			url: 'https://calendly.com',
			color: '#006BFF',
			category: 'business',
			subcategory: 'Scheduling & Documents',
			description: 'Meeting scheduling',
			logo: 'https://www.google.com/s2/favicons?domain=calendly.com&sz=128'
		},
		{
			name: 'DocuSign',
			url: 'https://app.docusign.com',
			color: '#FFCD00',
			category: 'business',
			subcategory: 'Scheduling & Documents',
			description: 'Electronic signatures',
			logo: 'https://www.google.com/s2/favicons?domain=docusign.com&sz=128'
		},
		{
			name: 'PandaDoc',
			url: 'https://app.pandadoc.com',
			color: '#59BC6A',
			category: 'business',
			subcategory: 'Scheduling & Documents',
			description: 'Document automation',
			logo: 'https://www.google.com/s2/favicons?domain=pandadoc.com&sz=128'
		},
		{
			name: 'Proposify',
			url: 'https://app.proposify.com',
			color: '#1BC5BD',
			category: 'business',
			subcategory: 'Scheduling & Documents',
			description: 'Proposal software',
			logo: 'https://www.google.com/s2/favicons?domain=proposify.com&sz=128'
		},
		// Business & CRM - Sales Intelligence
		{
			name: 'Apollo.io',
			url: 'https://app.apollo.io',
			color: '#6366F1',
			category: 'business',
			subcategory: 'Sales Intelligence',
			description: 'Sales intelligence',
			logo: 'https://www.google.com/s2/favicons?domain=apollo.io&sz=128'
		},
		{
			name: 'ZoomInfo',
			url: 'https://app.zoominfo.com',
			color: '#6A0DAD',
			category: 'business',
			subcategory: 'Sales Intelligence',
			description: 'B2B data platform',
			logo: 'https://www.google.com/s2/favicons?domain=zoominfo.com&sz=128'
		},
		{
			name: 'LinkedIn Sales',
			url: 'https://www.linkedin.com/sales/',
			color: '#0A66C2',
			category: 'business',
			subcategory: 'Sales Intelligence',
			description: 'Sales Navigator',
			logo: 'https://www.google.com/s2/favicons?domain=linkedin.com&sz=128'
		},
		{
			name: 'Outreach',
			url: 'https://app.outreach.io',
			color: '#5951FF',
			category: 'business',
			subcategory: 'Sales Intelligence',
			description: 'Sales engagement',
			logo: 'https://www.google.com/s2/favicons?domain=outreach.io&sz=128'
		},
		{
			name: 'Salesloft',
			url: 'https://app.salesloft.com',
			color: '#00B388',
			category: 'business',
			subcategory: 'Sales Intelligence',
			description: 'Sales engagement platform',
			logo: 'https://www.google.com/s2/favicons?domain=salesloft.com&sz=128'
		},
		{
			name: 'Gong',
			url: 'https://app.gong.io',
			color: '#7B61FF',
			category: 'business',
			subcategory: 'Sales Intelligence',
			description: 'Revenue intelligence',
			logo: 'https://www.google.com/s2/favicons?domain=gong.io&sz=128'
		},
		// Business & CRM - Marketing Automation
		{
			name: 'GoHighLevel',
			url: 'https://app.gohighlevel.com',
			color: '#00C6FF',
			category: 'business',
			subcategory: 'Marketing Automation',
			description: 'All-in-one marketing platform',
			logo: 'https://www.google.com/s2/favicons?domain=gohighlevel.com&sz=128',
			featured: true
		},
		{
			name: 'Keap',
			url: 'https://app.infusionsoft.com',
			color: '#159570',
			category: 'business',
			subcategory: 'Marketing Automation',
			description: 'CRM & automation',
			logo: 'https://www.google.com/s2/favicons?domain=keap.com&sz=128'
		},
		{
			name: 'ActiveCampaign',
			url: 'https://www.activecampaign.com',
			color: '#356AE6',
			category: 'business',
			subcategory: 'Marketing Automation',
			description: 'Email marketing & CRM',
			logo: 'https://www.google.com/s2/favicons?domain=activecampaign.com&sz=128'
		},
		{
			name: 'Mailchimp',
			url: 'https://mailchimp.com',
			color: '#FFE01B',
			category: 'business',
			subcategory: 'Marketing Automation',
			description: 'Email marketing',
			logo: 'https://www.google.com/s2/favicons?domain=mailchimp.com&sz=128'
		},
		{
			name: 'Klaviyo',
			url: 'https://www.klaviyo.com',
			color: '#111111',
			category: 'business',
			subcategory: 'Marketing Automation',
			description: 'E-commerce email marketing',
			logo: 'https://www.google.com/s2/favicons?domain=klaviyo.com&sz=128'
		},
		{
			name: 'Constant Contact',
			url: 'https://login.constantcontact.com',
			color: '#0070E0',
			category: 'business',
			subcategory: 'Marketing Automation',
			description: 'Email marketing',
			logo: 'https://www.google.com/s2/favicons?domain=constantcontact.com&sz=128'
		},
		{
			name: 'Brevo',
			url: 'https://app.brevo.com',
			color: '#0092FF',
			category: 'business',
			subcategory: 'Marketing Automation',
			description: 'Email & SMS marketing',
			logo: 'https://www.google.com/s2/favicons?domain=brevo.com&sz=128'
		},
		// Business & CRM - E-commerce
		{
			name: 'Shopify',
			url: 'https://admin.shopify.com',
			color: '#96BF48',
			category: 'business',
			subcategory: 'E-commerce',
			description: 'E-commerce platform',
			logo: 'https://www.google.com/s2/favicons?domain=shopify.com&sz=128'
		},
		{
			name: 'WooCommerce',
			url: 'https://woocommerce.com/my-dashboard/',
			color: '#96588A',
			category: 'business',
			subcategory: 'E-commerce',
			description: 'E-commerce for WordPress',
			logo: 'https://www.google.com/s2/favicons?domain=woocommerce.com&sz=128'
		},
		{
			name: 'BigCommerce',
			url: 'https://login.bigcommerce.com',
			color: '#121118',
			category: 'business',
			subcategory: 'E-commerce',
			description: 'E-commerce platform',
			logo: 'https://www.google.com/s2/favicons?domain=bigcommerce.com&sz=128'
		},
		// Business & CRM - Website Builders
		{
			name: 'Webflow',
			url: 'https://webflow.com/dashboard',
			color: '#4353FF',
			category: 'business',
			subcategory: 'Website Builders',
			description: 'Website builder',
			logo: 'https://www.google.com/s2/favicons?domain=webflow.com&sz=128'
		},
		{
			name: 'Squarespace',
			url: 'https://www.squarespace.com/config',
			color: '#000000',
			category: 'business',
			subcategory: 'Website Builders',
			description: 'Website builder',
			logo: 'https://www.google.com/s2/favicons?domain=squarespace.com&sz=128'
		},
		{
			name: 'Wix',
			url: 'https://manage.wix.com',
			color: '#0C6EFC',
			category: 'business',
			subcategory: 'Website Builders',
			description: 'Website builder',
			logo: 'https://www.google.com/s2/favicons?domain=wix.com&sz=128'
		},
		// Business & CRM - Landing Pages
		{
			name: 'ClickFunnels',
			url: 'https://app.clickfunnels.com',
			color: '#0066FF',
			category: 'business',
			subcategory: 'Landing Pages',
			description: 'Sales funnel builder',
			logo: 'https://www.google.com/s2/favicons?domain=clickfunnels.com&sz=128'
		},
		{
			name: 'Leadpages',
			url: 'https://my.leadpages.com',
			color: '#1656FF',
			category: 'business',
			subcategory: 'Landing Pages',
			description: 'Landing pages',
			logo: 'https://www.google.com/s2/favicons?domain=leadpages.com&sz=128'
		},
		{
			name: 'Unbounce',
			url: 'https://app.unbounce.com',
			color: '#1D4ED8',
			category: 'business',
			subcategory: 'Landing Pages',
			description: 'Landing page builder',
			logo: 'https://www.google.com/s2/favicons?domain=unbounce.com&sz=128'
		},
		// Business & CRM - Forms & Surveys
		{
			name: 'Typeform',
			url: 'https://admin.typeform.com',
			color: '#262627',
			category: 'business',
			subcategory: 'Forms & Surveys',
			description: 'Forms & surveys',
			logo: 'https://www.google.com/s2/favicons?domain=typeform.com&sz=128'
		},
		{
			name: 'JotForm',
			url: 'https://www.jotform.com/myforms/',
			color: '#FF6600',
			category: 'business',
			subcategory: 'Forms & Surveys',
			description: 'Online forms',
			logo: 'https://www.google.com/s2/favicons?domain=jotform.com&sz=128'
		},
		{
			name: 'Tally',
			url: 'https://tally.so',
			color: '#0D0D0D',
			category: 'business',
			subcategory: 'Forms & Surveys',
			description: 'Simple form builder',
			logo: 'https://www.google.com/s2/favicons?domain=tally.so&sz=128'
		},
		// Business & CRM - Analytics
		{
			name: 'Google Analytics',
			url: 'https://analytics.google.com',
			color: '#F9AB00',
			category: 'business',
			subcategory: 'Analytics',
			description: 'Web analytics',
			logo: 'https://www.google.com/s2/favicons?domain=analytics.google.com&sz=128'
		},
		{
			name: 'Mixpanel',
			url: 'https://mixpanel.com',
			color: '#7856FF',
			category: 'business',
			subcategory: 'Analytics',
			description: 'Product analytics',
			logo: 'https://www.google.com/s2/favicons?domain=mixpanel.com&sz=128'
		},
		{
			name: 'Amplitude',
			url: 'https://app.amplitude.com',
			color: '#0061F2',
			category: 'business',
			subcategory: 'Analytics',
			description: 'Product analytics',
			logo: 'https://www.google.com/s2/favicons?domain=amplitude.com&sz=128'
		},
		{
			name: 'Hotjar',
			url: 'https://insights.hotjar.com',
			color: '#FF3C00',
			category: 'business',
			subcategory: 'Analytics',
			description: 'User behavior analytics',
			logo: 'https://www.google.com/s2/favicons?domain=hotjar.com&sz=128'
		},
		// Business & CRM - SEO
		{
			name: 'SEMrush',
			url: 'https://www.semrush.com/dashboard/',
			color: '#FF642D',
			category: 'business',
			subcategory: 'SEO',
			description: 'SEO & marketing tools',
			logo: 'https://www.google.com/s2/favicons?domain=semrush.com&sz=128'
		},
		{
			name: 'Ahrefs',
			url: 'https://app.ahrefs.com',
			color: '#FF7139',
			category: 'business',
			subcategory: 'SEO',
			description: 'SEO toolset',
			logo: 'https://www.google.com/s2/favicons?domain=ahrefs.com&sz=128'
		},
		{
			name: 'Moz',
			url: 'https://moz.com/products',
			color: '#118FE3',
			category: 'business',
			subcategory: 'SEO',
			description: 'SEO software',
			logo: 'https://www.google.com/s2/favicons?domain=moz.com&sz=128'
		},
		// Business & CRM - Advertising
		{
			name: 'Google Ads',
			url: 'https://ads.google.com',
			color: '#4285F4',
			category: 'business',
			subcategory: 'Advertising',
			description: 'Online advertising',
			logo: 'https://www.google.com/s2/favicons?domain=ads.google.com&sz=128'
		},
		{
			name: 'Facebook Ads',
			url: 'https://business.facebook.com',
			color: '#1877F2',
			category: 'business',
			subcategory: 'Advertising',
			description: 'Social advertising',
			logo: 'https://www.google.com/s2/favicons?domain=facebook.com&sz=128'
		},
		// Business & CRM - HR & Payroll
		{
			name: 'BambooHR',
			url: 'https://app.bamboohr.com',
			color: '#73C41D',
			category: 'business',
			subcategory: 'HR & Payroll',
			description: 'HR software',
			logo: 'https://www.google.com/s2/favicons?domain=bamboohr.com&sz=128'
		},
		{
			name: 'Gusto',
			url: 'https://app.gusto.com',
			color: '#FF7155',
			category: 'business',
			subcategory: 'HR & Payroll',
			description: 'Payroll & HR',
			logo: 'https://www.google.com/s2/favicons?domain=gusto.com&sz=128'
		},
		{
			name: 'Rippling',
			url: 'https://app.rippling.com',
			color: '#FFC100',
			category: 'business',
			subcategory: 'HR & Payroll',
			description: 'HR & IT platform',
			logo: 'https://www.google.com/s2/favicons?domain=rippling.com&sz=128'
		},
		{
			name: 'Deel',
			url: 'https://app.deel.com',
			color: '#15357A',
			category: 'business',
			subcategory: 'HR & Payroll',
			description: 'Global payroll',
			logo: 'https://www.google.com/s2/favicons?domain=deel.com&sz=128'
		},
		// Business & CRM - Collaboration
		{
			name: 'Notion',
			url: 'https://notion.so',
			color: '#000000',
			category: 'business',
			subcategory: 'Collaboration',
			description: 'Company wiki & docs',
			logo: 'https://www.google.com/s2/favicons?domain=notion.so&sz=128'
		},
		{
			name: 'GitHub',
			url: 'https://github.com',
			color: '#24292E',
			category: 'business',
			subcategory: 'Collaboration',
			description: 'Code & project management',
			logo: 'https://www.google.com/s2/favicons?domain=github.com&sz=128'
		},
		// AI Tools - Chat Assistants
		{
			name: 'Claude',
			url: 'https://claude.ai',
			color: '#D97757',
			category: 'ai',
			subcategory: 'Chat Assistants',
			description: 'AI assistant by Anthropic',
			logo: 'https://www.google.com/s2/favicons?domain=claude.ai&sz=128',
			featured: true
		},
		{
			name: 'ChatGPT',
			url: 'https://chat.openai.com',
			color: '#10A37F',
			category: 'ai',
			subcategory: 'Chat Assistants',
			description: 'AI assistant by OpenAI',
			logo: '/logos/integrations/openai.svg',
			featured: true
		},
		{
			name: 'Perplexity',
			url: 'https://www.perplexity.ai',
			color: '#20808D',
			category: 'ai',
			subcategory: 'Chat Assistants',
			description: 'AI search engine',
			logo: 'https://www.google.com/s2/favicons?domain=perplexity.ai&sz=128',
			featured: true
		},
		{
			name: 'Gemini',
			url: 'https://gemini.google.com',
			color: '#8E75B2',
			category: 'ai',
			subcategory: 'Chat Assistants',
			description: 'AI assistant by Google',
			logo: 'https://www.google.com/s2/favicons?domain=gemini.google.com&sz=128'
		},
		{
			name: 'Copilot',
			url: 'https://copilot.microsoft.com',
			color: '#6264A7',
			category: 'ai',
			subcategory: 'Chat Assistants',
			description: 'AI assistant by Microsoft',
			logo: 'https://www.google.com/s2/favicons?domain=copilot.microsoft.com&sz=128'
		},
		{
			name: 'Mistral',
			url: 'https://chat.mistral.ai',
			color: '#FF7000',
			category: 'ai',
			subcategory: 'Chat Assistants',
			description: 'Open-source AI models',
			logo: 'https://www.google.com/s2/favicons?domain=mistral.ai&sz=128'
		},
		{
			name: 'Poe',
			url: 'https://poe.com',
			color: '#5B52E3',
			category: 'ai',
			subcategory: 'Chat Assistants',
			description: 'Multi-AI platform',
			logo: 'https://www.google.com/s2/favicons?domain=poe.com&sz=128'
		},
		// AI Tools - Writing & Content
		{
			name: 'Copy.ai',
			url: 'https://app.copy.ai',
			color: '#7C3AED',
			category: 'ai',
			subcategory: 'Writing & Content',
			description: 'AI copywriting',
			logo: 'https://www.google.com/s2/favicons?domain=copy.ai&sz=128'
		},
		{
			name: 'Jasper',
			url: 'https://app.jasper.ai',
			color: '#FF6B6B',
			category: 'ai',
			subcategory: 'Writing & Content',
			description: 'AI content creation',
			logo: 'https://www.google.com/s2/favicons?domain=jasper.ai&sz=128'
		},
		{
			name: 'Writesonic',
			url: 'https://app.writesonic.com',
			color: '#5B5FC7',
			category: 'ai',
			subcategory: 'Writing & Content',
			description: 'AI writing assistant',
			logo: 'https://www.google.com/s2/favicons?domain=writesonic.com&sz=128'
		},
		{
			name: 'Grammarly',
			url: 'https://app.grammarly.com',
			color: '#15C39A',
			category: 'ai',
			subcategory: 'Writing & Content',
			description: 'AI writing assistant',
			logo: 'https://www.google.com/s2/favicons?domain=grammarly.com&sz=128'
		},
		{
			name: 'Wordtune',
			url: 'https://www.wordtune.com',
			color: '#FF0064',
			category: 'ai',
			subcategory: 'Writing & Content',
			description: 'AI writing rewriter',
			logo: 'https://www.google.com/s2/favicons?domain=wordtune.com&sz=128'
		},
		{
			name: 'Notion AI',
			url: 'https://www.notion.so/product/ai',
			color: '#000000',
			category: 'ai',
			subcategory: 'Writing & Content',
			description: 'AI in your workspace',
			logo: 'https://www.google.com/s2/favicons?domain=notion.so&sz=128'
		},
		// AI Tools - Meeting & Transcription
		{
			name: 'Otter.ai',
			url: 'https://otter.ai',
			color: '#0047FF',
			category: 'ai',
			subcategory: 'Meeting & Transcription',
			description: 'AI meeting notes',
			logo: 'https://www.google.com/s2/favicons?domain=otter.ai&sz=128'
		},
		{
			name: 'Fireflies.ai',
			url: 'https://app.fireflies.ai',
			color: '#B429F9',
			category: 'ai',
			subcategory: 'Meeting & Transcription',
			description: 'AI meeting assistant',
			logo: 'https://www.google.com/s2/favicons?domain=fireflies.ai&sz=128'
		},
		{
			name: 'Fathom',
			url: 'https://fathom.video',
			color: '#5046E5',
			category: 'ai',
			subcategory: 'Meeting & Transcription',
			description: 'AI meeting recorder',
			logo: 'https://www.google.com/s2/favicons?domain=fathom.video&sz=128'
		},
		{
			name: 'Grain',
			url: 'https://grain.com',
			color: '#FF6B6B',
			category: 'ai',
			subcategory: 'Meeting & Transcription',
			description: 'Meeting highlights',
			logo: 'https://www.google.com/s2/favicons?domain=grain.com&sz=128'
		},
		{
			name: 'tl;dv',
			url: 'https://tldv.io',
			color: '#7C3AED',
			category: 'ai',
			subcategory: 'Meeting & Transcription',
			description: 'Meeting recorder & summary',
			logo: 'https://www.google.com/s2/favicons?domain=tldv.io&sz=128'
		},
		// AI Tools - Image Generation
		{
			name: 'Midjourney',
			url: 'https://www.midjourney.com',
			color: '#000000',
			category: 'ai',
			subcategory: 'Image Generation',
			description: 'AI image generation',
			logo: 'https://www.google.com/s2/favicons?domain=midjourney.com&sz=128'
		},
		{
			name: 'DALL-E',
			url: 'https://labs.openai.com',
			color: '#10A37F',
			category: 'ai',
			subcategory: 'Image Generation',
			description: 'AI image generation',
			logo: 'https://www.google.com/s2/favicons?domain=openai.com&sz=128'
		},
		{
			name: 'Stable Diffusion',
			url: 'https://stability.ai',
			color: '#7C3AED',
			category: 'ai',
			subcategory: 'Image Generation',
			description: 'AI image generation',
			logo: 'https://www.google.com/s2/favicons?domain=stability.ai&sz=128'
		},
		{
			name: 'Leonardo.ai',
			url: 'https://app.leonardo.ai',
			color: '#7E22CE',
			category: 'ai',
			subcategory: 'Image Generation',
			description: 'AI art generator',
			logo: 'https://www.google.com/s2/favicons?domain=leonardo.ai&sz=128'
		},
		// AI Tools - Video Generation
		{
			name: 'Runway',
			url: 'https://app.runwayml.com',
			color: '#000000',
			category: 'ai',
			subcategory: 'Video Generation',
			description: 'AI video tools',
			logo: 'https://www.google.com/s2/favicons?domain=runwayml.com&sz=128'
		},
		{
			name: 'Pika',
			url: 'https://pika.art',
			color: '#FF5C00',
			category: 'ai',
			subcategory: 'Video Generation',
			description: 'AI video generation',
			logo: 'https://www.google.com/s2/favicons?domain=pika.art&sz=128'
		},
		{
			name: 'Synthesia',
			url: 'https://www.synthesia.io',
			color: '#6C3BFF',
			category: 'ai',
			subcategory: 'Video Generation',
			description: 'AI video generation',
			logo: 'https://www.google.com/s2/favicons?domain=synthesia.io&sz=128'
		},
		{
			name: 'HeyGen',
			url: 'https://app.heygen.com',
			color: '#2B5BFF',
			category: 'ai',
			subcategory: 'Video Generation',
			description: 'AI avatar videos',
			logo: 'https://www.google.com/s2/favicons?domain=heygen.com&sz=128'
		},
		{
			name: 'Descript',
			url: 'https://www.descript.com',
			color: '#00E676',
			category: 'ai',
			subcategory: 'Video Generation',
			description: 'AI video/audio editing',
			logo: 'https://www.google.com/s2/favicons?domain=descript.com&sz=128'
		},
		{
			name: 'Higgsfield',
			url: 'https://higgsfield.ai',
			color: '#7C3AED',
			category: 'ai',
			subcategory: 'Video Generation',
			description: 'AI video creation',
			logo: 'https://www.google.com/s2/favicons?domain=higgsfield.ai&sz=128'
		},
		{
			name: 'Sora',
			url: 'https://openai.com/sora',
			color: '#10A37F',
			category: 'ai',
			subcategory: 'Video Generation',
			description: 'AI video generation',
			logo: 'https://www.google.com/s2/favicons?domain=openai.com&sz=128'
		},
		{
			name: 'Kling AI',
			url: 'https://klingai.com',
			color: '#FF6B35',
			category: 'ai',
			subcategory: 'Video Generation',
			description: 'AI video generation',
			logo: 'https://www.google.com/s2/favicons?domain=klingai.com&sz=128'
		},
		// AI Tools - Audio & Voice
		{
			name: 'ElevenLabs',
			url: 'https://elevenlabs.io',
			color: '#000000',
			category: 'ai',
			subcategory: 'Audio & Voice',
			description: 'AI voice generation',
			logo: 'https://www.google.com/s2/favicons?domain=elevenlabs.io&sz=128'
		},
		{
			name: 'Suno',
			url: 'https://suno.ai',
			color: '#000000',
			category: 'ai',
			subcategory: 'Audio & Voice',
			description: 'AI music generation',
			logo: 'https://www.google.com/s2/favicons?domain=suno.ai&sz=128'
		},
		{
			name: 'Udio',
			url: 'https://www.udio.com',
			color: '#FF3366',
			category: 'ai',
			subcategory: 'Audio & Voice',
			description: 'AI music creation',
			logo: 'https://www.google.com/s2/favicons?domain=udio.com&sz=128'
		},
		// AI Tools - Data & Research
		{
			name: 'Explorium',
			url: 'https://www.explorium.ai',
			color: '#4F46E5',
			category: 'ai',
			subcategory: 'Data & Research',
			description: 'AI data enrichment',
			logo: 'https://www.google.com/s2/favicons?domain=explorium.ai&sz=128'
		},
		// Productivity - Documents
		{
			name: 'Google Docs',
			url: 'https://docs.google.com',
			color: '#4285F4',
			category: 'productivity',
			subcategory: 'Documents',
			description: 'Document editing',
			logo: 'https://www.google.com/s2/favicons?domain=docs.google.com&sz=128'
		},
		{
			name: 'Microsoft Word',
			url: 'https://www.office.com/launch/word',
			color: '#2B579A',
			category: 'productivity',
			subcategory: 'Documents',
			description: 'Document editor',
			logo: 'https://www.google.com/s2/favicons?domain=office.com&sz=128'
		},
		{
			name: 'Coda',
			url: 'https://coda.io',
			color: '#F46A54',
			category: 'productivity',
			subcategory: 'Documents',
			description: 'Doc meets spreadsheet',
			logo: 'https://www.google.com/s2/favicons?domain=coda.io&sz=128'
		},
		// Productivity - Spreadsheets
		{
			name: 'Google Sheets',
			url: 'https://sheets.google.com',
			color: '#0F9D58',
			category: 'productivity',
			subcategory: 'Spreadsheets',
			description: 'Spreadsheets',
			logo: 'https://www.google.com/s2/favicons?domain=sheets.google.com&sz=128'
		},
		{
			name: 'Microsoft Excel',
			url: 'https://www.office.com/launch/excel',
			color: '#217346',
			category: 'productivity',
			subcategory: 'Spreadsheets',
			description: 'Spreadsheet editor',
			logo: 'https://www.google.com/s2/favicons?domain=office.com&sz=128'
		},
		{
			name: 'Airtable',
			url: 'https://airtable.com',
			color: '#18BFFF',
			category: 'productivity',
			subcategory: 'Spreadsheets',
			description: 'Database & spreadsheet',
			logo: 'https://www.google.com/s2/favicons?domain=airtable.com&sz=128'
		},
		// Productivity - Notes & Knowledge
		{
			name: 'Evernote',
			url: 'https://www.evernote.com',
			color: '#00A82D',
			category: 'productivity',
			subcategory: 'Notes & Knowledge',
			description: 'Note-taking app',
			logo: 'https://www.google.com/s2/favicons?domain=evernote.com&sz=128'
		},
		{
			name: 'Obsidian',
			url: 'https://obsidian.md',
			color: '#7C3AED',
			category: 'productivity',
			subcategory: 'Notes & Knowledge',
			description: 'Knowledge base',
			logo: 'https://www.google.com/s2/favicons?domain=obsidian.md&sz=128'
		},
		{
			name: 'Roam Research',
			url: 'https://roamresearch.com',
			color: '#154B78',
			category: 'productivity',
			subcategory: 'Notes & Knowledge',
			description: 'Networked thought',
			logo: 'https://www.google.com/s2/favicons?domain=roamresearch.com&sz=128'
		},
		// Productivity - Calendar & Scheduling
		{
			name: 'Google Calendar',
			url: 'https://calendar.google.com',
			color: '#4285F4',
			category: 'productivity',
			subcategory: 'Calendar & Scheduling',
			description: 'Calendar & scheduling',
			logo: 'https://www.google.com/s2/favicons?domain=calendar.google.com&sz=128'
		},
		// Productivity - Task Management
		{
			name: 'Todoist',
			url: 'https://todoist.com',
			color: '#E44332',
			category: 'productivity',
			subcategory: 'Task Management',
			description: 'Task management',
			logo: 'https://www.google.com/s2/favicons?domain=todoist.com&sz=128'
		},
		{
			name: 'TickTick',
			url: 'https://ticktick.com',
			color: '#4772FA',
			category: 'productivity',
			subcategory: 'Task Management',
			description: 'Task manager',
			logo: 'https://www.google.com/s2/favicons?domain=ticktick.com&sz=128'
		},
		{
			name: 'Any.do',
			url: 'https://www.any.do',
			color: '#4C84FF',
			category: 'productivity',
			subcategory: 'Task Management',
			description: 'To-do list & planner',
			logo: 'https://www.google.com/s2/favicons?domain=any.do&sz=128'
		},
		// Project Management - Agile & Issues
		{
			name: 'Linear',
			url: 'https://linear.app',
			color: '#5E6AD2',
			category: 'project-management',
			subcategory: 'Agile & Issues',
			description: 'Issue tracking',
			logo: 'https://www.google.com/s2/favicons?domain=linear.app&sz=128',
			featured: true
		},
		{
			name: 'Jira',
			url: 'https://atlassian.net',
			color: '#0052CC',
			category: 'project-management',
			subcategory: 'Agile & Issues',
			description: 'Issue & project tracking',
			logo: 'https://www.google.com/s2/favicons?domain=atlassian.net&sz=128'
		},
		{
			name: 'Height',
			url: 'https://height.app',
			color: '#6366F1',
			category: 'project-management',
			subcategory: 'Agile & Issues',
			description: 'Team collaboration',
			logo: 'https://www.google.com/s2/favicons?domain=height.app&sz=128'
		},
		// Project Management - Work Management
		{
			name: 'Asana',
			url: 'https://app.asana.com',
			color: '#F06A6A',
			category: 'project-management',
			subcategory: 'Work Management',
			description: 'Work management',
			logo: 'https://www.google.com/s2/favicons?domain=asana.com&sz=128'
		},
		{
			name: 'ClickUp',
			url: 'https://app.clickup.com',
			color: '#7B68EE',
			category: 'project-management',
			subcategory: 'Work Management',
			description: 'Project management',
			logo: 'https://www.google.com/s2/favicons?domain=clickup.com&sz=128'
		},
		{
			name: 'Monday',
			url: 'https://monday.com',
			color: '#FF3D57',
			category: 'project-management',
			subcategory: 'Work Management',
			description: 'Work OS',
			logo: 'https://www.google.com/s2/favicons?domain=monday.com&sz=128'
		},
		{
			name: 'Wrike',
			url: 'https://www.wrike.com',
			color: '#2CD598',
			category: 'project-management',
			subcategory: 'Work Management',
			description: 'Work management',
			logo: 'https://www.google.com/s2/favicons?domain=wrike.com&sz=128'
		},
		{
			name: 'Teamwork',
			url: 'https://www.teamwork.com',
			color: '#6C63FF',
			category: 'project-management',
			subcategory: 'Work Management',
			description: 'Project management',
			logo: 'https://www.google.com/s2/favicons?domain=teamwork.com&sz=128'
		},
		{
			name: 'Smartsheet',
			url: 'https://app.smartsheet.com',
			color: '#0E7A0E',
			category: 'project-management',
			subcategory: 'Work Management',
			description: 'Work automation',
			logo: 'https://www.google.com/s2/favicons?domain=smartsheet.com&sz=128'
		},
		{
			name: 'Basecamp',
			url: 'https://basecamp.com',
			color: '#1D2D35',
			category: 'project-management',
			subcategory: 'Work Management',
			description: 'Project management',
			logo: 'https://www.google.com/s2/favicons?domain=basecamp.com&sz=128'
		},
		{
			name: 'Notion Projects',
			url: 'https://notion.so',
			color: '#000000',
			category: 'project-management',
			subcategory: 'Work Management',
			description: 'Project workspace',
			logo: 'https://www.google.com/s2/favicons?domain=notion.so&sz=128'
		},
		// Project Management - Kanban
		{
			name: 'Trello',
			url: 'https://trello.com',
			color: '#0079BF',
			category: 'project-management',
			subcategory: 'Kanban',
			description: 'Kanban boards',
			logo: 'https://www.google.com/s2/favicons?domain=trello.com&sz=128'
		},
		// Communication - Team Messaging
		{
			name: 'Slack',
			url: 'https://app.slack.com',
			color: '#4A154B',
			category: 'communication',
			subcategory: 'Team Messaging',
			description: 'Team messaging',
			logo: 'https://www.google.com/s2/favicons?domain=slack.com&sz=128',
			featured: true
		},
		{
			name: 'Discord',
			url: 'https://discord.com/app',
			color: '#5865F2',
			category: 'communication',
			subcategory: 'Team Messaging',
			description: 'Voice & text chat',
			logo: 'https://www.google.com/s2/favicons?domain=discord.com&sz=128'
		},
		{
			name: 'Microsoft Teams',
			url: 'https://teams.microsoft.com',
			color: '#6264A7',
			category: 'communication',
			subcategory: 'Team Messaging',
			description: 'Team collaboration',
			logo: 'https://www.google.com/s2/favicons?domain=teams.microsoft.com&sz=128'
		},
		// Communication - Video Conferencing
		{
			name: 'Zoom',
			url: 'https://zoom.us/meeting',
			color: '#2D8CFF',
			category: 'communication',
			subcategory: 'Video Conferencing',
			description: 'Video meetings',
			logo: 'https://www.google.com/s2/favicons?domain=zoom.us&sz=128'
		},
		{
			name: 'Google Meet',
			url: 'https://meet.google.com',
			color: '#00897B',
			category: 'communication',
			subcategory: 'Video Conferencing',
			description: 'Video meetings',
			logo: 'https://www.google.com/s2/favicons?domain=meet.google.com&sz=128'
		},
		{
			name: 'Webex',
			url: 'https://web.webex.com',
			color: '#00B140',
			category: 'communication',
			subcategory: 'Video Conferencing',
			description: 'Video conferencing',
			logo: 'https://www.google.com/s2/favicons?domain=webex.com&sz=128'
		},
		{
			name: 'Around',
			url: 'https://www.around.co',
			color: '#6C5CE7',
			category: 'communication',
			subcategory: 'Video Conferencing',
			description: 'Video collaboration',
			logo: 'https://www.google.com/s2/favicons?domain=around.co&sz=128'
		},
		{
			name: 'Loom',
			url: 'https://www.loom.com',
			color: '#625DF5',
			category: 'communication',
			subcategory: 'Video Conferencing',
			description: 'Video messages',
			logo: 'https://www.google.com/s2/favicons?domain=loom.com&sz=128'
		},
		// Communication - Email
		{
			name: 'Gmail',
			url: 'https://mail.google.com',
			color: '#EA4335',
			category: 'communication',
			subcategory: 'Email',
			description: 'Email by Google',
			logo: 'https://www.google.com/s2/favicons?domain=mail.google.com&sz=128'
		},
		{
			name: 'Outlook',
			url: 'https://outlook.live.com',
			color: '#0078D4',
			category: 'communication',
			subcategory: 'Email',
			description: 'Email by Microsoft',
			logo: 'https://www.google.com/s2/favicons?domain=outlook.live.com&sz=128'
		},
		// Communication - Personal Messaging
		{
			name: 'WhatsApp Web',
			url: 'https://web.whatsapp.com',
			color: '#25D366',
			category: 'communication',
			subcategory: 'Personal Messaging',
			description: 'Messaging app',
			logo: 'https://www.google.com/s2/favicons?domain=whatsapp.com&sz=128'
		},
		{
			name: 'Telegram',
			url: 'https://web.telegram.org',
			color: '#0088CC',
			category: 'communication',
			subcategory: 'Personal Messaging',
			description: 'Secure messaging',
			logo: 'https://www.google.com/s2/favicons?domain=telegram.org&sz=128'
		},
		{
			name: 'Signal',
			url: 'https://signal.org',
			color: '#3A76F0',
			category: 'communication',
			subcategory: 'Personal Messaging',
			description: 'Private messaging',
			logo: 'https://www.google.com/s2/favicons?domain=signal.org&sz=128'
		},
		// Design - UI/UX Design
		{
			name: 'Figma',
			url: 'https://www.figma.com',
			color: '#F24E1E',
			category: 'design',
			subcategory: 'UI/UX Design',
			description: 'Design tool',
			logo: 'https://www.google.com/s2/favicons?domain=figma.com&sz=128',
			featured: true
		},
		{
			name: 'Framer',
			url: 'https://www.framer.com',
			color: '#0055FF',
			category: 'design',
			subcategory: 'UI/UX Design',
			description: 'Interactive design',
			logo: 'https://www.google.com/s2/favicons?domain=framer.com&sz=128'
		},
		{
			name: 'Sketch',
			url: 'https://www.sketch.com',
			color: '#F7B500',
			category: 'design',
			subcategory: 'UI/UX Design',
			description: 'Digital design',
			logo: 'https://www.google.com/s2/favicons?domain=sketch.com&sz=128'
		},
		// Design - Graphic Design
		{
			name: 'Canva',
			url: 'https://www.canva.com',
			color: '#00C4CC',
			category: 'design',
			subcategory: 'Graphic Design',
			description: 'Graphic design',
			logo: 'https://www.google.com/s2/favicons?domain=canva.com&sz=128'
		},
		{
			name: 'Adobe Creative Cloud',
			url: 'https://www.adobe.com/creativecloud.html',
			color: '#FF0000',
			category: 'design',
			subcategory: 'Graphic Design',
			description: 'Creative suite',
			logo: 'https://www.google.com/s2/favicons?domain=adobe.com&sz=128'
		},
		{
			name: 'Photoshop',
			url: 'https://photoshop.adobe.com',
			color: '#31A8FF',
			category: 'design',
			subcategory: 'Photo Editing',
			description: 'Photo editing & graphics',
			logo: 'https://www.google.com/s2/favicons?domain=photoshop.adobe.com&sz=128',
			featured: true
		},
		{
			name: 'Lightroom',
			url: 'https://lightroom.adobe.com',
			color: '#31A8FF',
			category: 'design',
			subcategory: 'Photo Editing',
			description: 'Photo editing & organization',
			logo: 'https://www.google.com/s2/favicons?domain=lightroom.adobe.com&sz=128'
		},
		{
			name: 'After Effects',
			url: 'https://www.adobe.com/products/aftereffects.html',
			color: '#9999FF',
			category: 'design',
			subcategory: 'Motion Graphics',
			description: 'Motion graphics & VFX',
			logo: 'https://www.google.com/s2/favicons?domain=adobe.com&sz=128'
		},
		{
			name: 'Premiere Pro',
			url: 'https://www.adobe.com/products/premiere.html',
			color: '#9999FF',
			category: 'design',
			subcategory: 'Video Editing',
			description: 'Professional video editing',
			logo: 'https://www.google.com/s2/favicons?domain=adobe.com&sz=128'
		},
		{
			name: 'Illustrator',
			url: 'https://www.adobe.com/products/illustrator.html',
			color: '#FF9A00',
			category: 'design',
			subcategory: 'Vector Graphics',
			description: 'Vector graphics & illustration',
			logo: 'https://www.google.com/s2/favicons?domain=adobe.com&sz=128'
		},
		// Design - Video & Animation
		{
			name: 'CapCut',
			url: 'https://www.capcut.com',
			color: '#000000',
			category: 'design',
			subcategory: 'Video Editing',
			description: 'Video editing app',
			logo: 'https://www.google.com/s2/favicons?domain=capcut.com&sz=128',
			featured: true
		},
		{
			name: 'Rive',
			url: 'https://rive.app',
			color: '#1D1D1D',
			category: 'design',
			subcategory: 'Motion Graphics',
			description: 'Interactive animations',
			logo: 'https://www.google.com/s2/favicons?domain=rive.app&sz=128',
			featured: true
		},
		{
			name: 'Lottie',
			url: 'https://lottiefiles.com',
			color: '#00DDB3',
			category: 'design',
			subcategory: 'Motion Graphics',
			description: 'Animation library',
			logo: 'https://www.google.com/s2/favicons?domain=lottiefiles.com&sz=128'
		},
		{
			name: 'Spline',
			url: 'https://spline.design',
			color: '#7B61FF',
			category: 'design',
			subcategory: '3D Design',
			description: '3D design for web',
			logo: 'https://www.google.com/s2/favicons?domain=spline.design&sz=128'
		},
		{
			name: 'Unicorn Studio',
			url: 'https://www.unicorn.studio',
			color: '#FF6B6B',
			category: 'design',
			subcategory: 'Motion Graphics',
			description: 'WebGL visual effects',
			logo: 'https://www.google.com/s2/favicons?domain=unicorn.studio&sz=128',
			featured: true
		},
		{
			name: 'Blender',
			url: 'https://www.blender.org',
			color: '#F5792A',
			category: 'design',
			subcategory: '3D Design',
			description: '3D creation suite',
			logo: 'https://www.google.com/s2/favicons?domain=blender.org&sz=128'
		},
		{
			name: 'Cinema 4D',
			url: 'https://www.maxon.net/cinema-4d',
			color: '#011A37',
			category: 'design',
			subcategory: '3D Design',
			description: 'Professional 3D',
			logo: 'https://www.google.com/s2/favicons?domain=maxon.net&sz=128'
		},
		// Design - Whiteboard
		{
			name: 'Miro',
			url: 'https://miro.com',
			color: '#FFD02F',
			category: 'design',
			subcategory: 'Whiteboard',
			description: 'Whiteboard',
			logo: 'https://www.google.com/s2/favicons?domain=miro.com&sz=128'
		},
		{
			name: 'FigJam',
			url: 'https://www.figma.com/figjam/',
			color: '#F24E1E',
			category: 'design',
			subcategory: 'Whiteboard',
			description: 'Collaborative whiteboard',
			logo: 'https://www.google.com/s2/favicons?domain=figma.com&sz=128'
		},
		{
			name: 'Whimsical',
			url: 'https://whimsical.com',
			color: '#7C3AED',
			category: 'design',
			subcategory: 'Whiteboard',
			description: 'Visual workspace',
			logo: 'https://www.google.com/s2/favicons?domain=whimsical.com&sz=128'
		},
		// Storage - Cloud Storage
		{
			name: 'Google Drive',
			url: 'https://drive.google.com',
			color: '#4285F4',
			category: 'storage',
			subcategory: 'Cloud Storage',
			description: 'Cloud storage',
			logo: 'https://www.google.com/s2/favicons?domain=drive.google.com&sz=128'
		},
		{
			name: 'Dropbox',
			url: 'https://www.dropbox.com',
			color: '#0061FF',
			category: 'storage',
			subcategory: 'Cloud Storage',
			description: 'File storage',
			logo: 'https://www.google.com/s2/favicons?domain=dropbox.com&sz=128'
		},
		{
			name: 'OneDrive',
			url: 'https://onedrive.live.com',
			color: '#0078D4',
			category: 'storage',
			subcategory: 'Cloud Storage',
			description: 'Cloud storage',
			logo: 'https://www.google.com/s2/favicons?domain=onedrive.live.com&sz=128'
		},
		{
			name: 'Box',
			url: 'https://www.box.com',
			color: '#0061D5',
			category: 'storage',
			subcategory: 'Cloud Storage',
			description: 'Enterprise storage',
			logo: 'https://www.google.com/s2/favicons?domain=box.com&sz=128'
		},
		// Media - Video
		{
			name: 'YouTube',
			url: 'https://www.youtube.com',
			color: '#FF0000',
			category: 'media',
			subcategory: 'Video',
			description: 'Video streaming',
			logo: 'https://www.google.com/s2/favicons?domain=youtube.com&sz=128'
		},
		{
			name: 'YouTube Studio',
			url: 'https://studio.youtube.com',
			color: '#FF0000',
			category: 'media',
			subcategory: 'Video',
			description: 'Video management',
			logo: 'https://www.google.com/s2/favicons?domain=youtube.com&sz=128'
		},
		{
			name: 'Vimeo',
			url: 'https://vimeo.com',
			color: '#1AB7EA',
			category: 'media',
			subcategory: 'Video',
			description: 'Video hosting',
			logo: 'https://www.google.com/s2/favicons?domain=vimeo.com&sz=128'
		},
		// Media - Audio
		{
			name: 'Spotify',
			url: 'https://open.spotify.com',
			color: '#1DB954',
			category: 'media',
			subcategory: 'Audio',
			description: 'Music streaming',
			logo: 'https://www.google.com/s2/favicons?domain=open.spotify.com&sz=128'
		},
		{
			name: 'Apple Music',
			url: 'https://music.apple.com',
			color: '#FC3C44',
			category: 'media',
			subcategory: 'Audio',
			description: 'Music streaming',
			logo: 'https://www.google.com/s2/favicons?domain=music.apple.com&sz=128'
		},
		// Social - Professional
		{
			name: 'LinkedIn',
			url: 'https://www.linkedin.com',
			color: '#0A66C2',
			category: 'social',
			subcategory: 'Professional',
			description: 'Professional network',
			logo: 'https://www.google.com/s2/favicons?domain=linkedin.com&sz=128'
		},
		// Social - Social Networks
		{
			name: 'X (Twitter)',
			url: 'https://x.com',
			color: '#000000',
			category: 'social',
			subcategory: 'Social Networks',
			description: 'Social network',
			logo: 'https://www.google.com/s2/favicons?domain=x.com&sz=128'
		},
		{
			name: 'Facebook',
			url: 'https://www.facebook.com',
			color: '#1877F2',
			category: 'social',
			subcategory: 'Social Networks',
			description: 'Social network',
			logo: 'https://www.google.com/s2/favicons?domain=facebook.com&sz=128'
		},
		{
			name: 'Instagram',
			url: 'https://www.instagram.com',
			color: '#E4405F',
			category: 'social',
			subcategory: 'Social Networks',
			description: 'Photo & video sharing',
			logo: 'https://www.google.com/s2/favicons?domain=instagram.com&sz=128'
		},
		{
			name: 'TikTok',
			url: 'https://www.tiktok.com',
			color: '#000000',
			category: 'social',
			subcategory: 'Social Networks',
			description: 'Short-form video',
			logo: 'https://www.google.com/s2/favicons?domain=tiktok.com&sz=128'
		},
		{
			name: 'Reddit',
			url: 'https://www.reddit.com',
			color: '#FF4500',
			category: 'social',
			subcategory: 'Social Networks',
			description: 'Community forums',
			logo: 'https://www.google.com/s2/favicons?domain=reddit.com&sz=128'
		}
	];

	// Featured apps
	const featuredApps = $derived(popularApps.filter((app) => app.featured));

	// Filter apps by category and search
	const filteredApps = $derived(
		popularApps.filter((app) => {
			const matchesCategory = selectedCategory === 'all' || app.category === selectedCategory;
			const matchesSearch =
				!searchQuery ||
				app.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
				app.description.toLowerCase().includes(searchQuery.toLowerCase());
			return matchesCategory && matchesSearch;
		})
	);

	// Group apps by subcategory (only when viewing a specific category, not "all")
	const groupedApps = $derived(() => {
		if (selectedCategory === 'all' || searchQuery) {
			// Don't group when showing all or searching
			return [{ subcategory: null, apps: filteredApps }];
		}

		// Group by subcategory
		const groups: Record<string, typeof filteredApps> = {};
		const noSubcategory: typeof filteredApps = [];

		for (const app of filteredApps) {
			if (app.subcategory) {
				if (!groups[app.subcategory]) {
					groups[app.subcategory] = [];
				}
				groups[app.subcategory].push(app);
			} else {
				noSubcategory.push(app);
			}
		}

		const result: { subcategory: string | null; apps: typeof filteredApps }[] = [];

		// Add grouped apps first
		for (const [subcategory, apps] of Object.entries(groups)) {
			result.push({ subcategory, apps });
		}

		// Add ungrouped apps at the end
		if (noSubcategory.length > 0) {
			result.push({ subcategory: null, apps: noSubcategory });
		}

		return result;
	});

	// Get category count
	function getCategoryCount(categoryId: string): number {
		if (categoryId === 'all') return popularApps.length;
		return popularApps.filter((app) => app.category === categoryId).length;
	}

	// Check if app is already added (uses derived array for proper reactivity)
	function isAppAdded(appUrl: string): boolean {
		return addedAppUrls.includes(appUrl);
	}

	// Get the user app by URL (for removing)
	function getUserAppByUrl(appUrl: string) {
		return $userAppsStore.apps.find((app) => app.url === appUrl);
	}

	// Handle app card click - add if not added, remove if already added
	async function handleAppClick(app: (typeof popularApps)[0]) {
		const existingApp = getUserAppByUrl(app.url);
		if (existingApp) {
			// App is already added - remove it
			try {
				await userAppsStore.delete(existingApp.id, workspaceId);
				desktop3dStore.removeUserApp(existingApp.id);
			} catch (error) {
				console.error('Failed to remove app:', error);
			}
		} else {
			// App not added - add it
			await quickAdd(app);
		}
	}

	// Fetch user apps on mount
	$effect(() => {
		userAppsStore.fetch(workspaceId);
	});

	async function quickAdd(app: (typeof popularApps)[0]) {
		if (isAppAdded(app.url)) return;

		isSubmitting = true;
		try {
			const params: CreateUserAppParams = {
				workspace_id: workspaceId,
				name: app.name,
				url: app.url,
				color: app.color,
				logo_url: app.logo,
				category: app.category,
				description: app.description,
				app_type: 'web'
			};

			const newApp = await userAppsStore.create(params);

			if (newApp) {
				desktop3dStore.addUserApp(newApp);
			}

			dispatch('appCreated');
		} catch (error) {
			console.error('Failed to add app:', error);
			alert('Failed to add app. Please try again.');
		} finally {
			isSubmitting = false;
		}
	}

	async function handleCustomSubmit(e: Event) {
		e.preventDefault();

		if (!name || !url) {
			return;
		}

		isSubmitting = true;

		try {
			const params: CreateUserAppParams = {
				workspace_id: workspaceId,
				name,
				url,
				color: '#6366F1',
				category,
				description: description || undefined,
				app_type: 'web'
			};

			const newApp = await userAppsStore.create(params);

			if (newApp) {
				desktop3dStore.addUserApp(newApp);
			}

			name = '';
			url = '';
			category = 'productivity';
			description = '';

			dispatch('appCreated');
			onClose?.();
		} catch (error) {
			console.error('Failed to create app:', error);
			alert('Failed to create app. Please try again.');
		} finally {
			isSubmitting = false;
		}
	}

	async function deleteApp(appId: string) {
		if (confirm('Remove this app?')) {
			try {
				await userAppsStore.delete(appId, workspaceId);
				desktop3dStore.removeUserApp(appId);
			} catch (error) {
				console.error('Failed to delete app:', error);
				alert('Failed to delete app. Please try again.');
			}
		}
	}

	async function toggleAppActive(app: { id: string; is_active?: boolean | null }) {
		const newState = app.is_active === false ? true : false;
		try {
			await userAppsStore.update(app.id, workspaceId, { is_active: newState });
			// Note: Desktop3D will update automatically through store subscription
		} catch (error) {
			console.error('Failed to toggle app:', error);
		}
	}

	async function removeApp(appId: string) {
		if (confirm('Remove this app from your desktop?')) {
			try {
				await userAppsStore.delete(appId, workspaceId);
				desktop3dStore.removeUserApp(appId);
			} catch (error) {
				console.error('Failed to remove app:', error);
				alert('Failed to remove app. Please try again.');
			}
		}
	}
</script>

<!-- Wrapper div handles both page and modal mode -->
<div class={isPage ? 'page-container' : 'modal-overlay'} role={isPage ? undefined : 'dialog'} aria-modal={isPage ? undefined : 'true'}>
	<div class={isPage ? 'page-content' : 'modal-container'}>
		<!-- Header - only for modal mode -->
		{#if !isPage}
			<div class="modal-header">
				<div class="header-content">
					<div class="header-text">
						<h2>App Store</h2>
						<p>Add your favorite apps to BusinessOS</p>
					</div>
				</div>
				{#if onClose}
					<button class="close-btn" onclick={onClose} aria-label="Close">
						<X size={20} />
					</button>
				{/if}
			</div>
		{/if}

		<!-- Main Tabs -->
		<div class="btn-pill-group main-tabs">
			<button
				class="btn-pill btn-pill-sm {activeTab === 'browse' ? 'btn-pill-primary' : 'btn-pill-ghost'}"
				onclick={() => (activeTab = 'browse')}
			>
				<Globe size={16} />
				Browse Apps
			</button>
			<button
				class="btn-pill btn-pill-sm {activeTab === 'myapps' ? 'btn-pill-primary' : 'btn-pill-ghost'}"
				onclick={() => (activeTab = 'myapps')}
			>
				<Layers size={16} />
				My Apps
				{#if $userAppsStore.apps.length > 0}
					<span class="tab-badge">{$userAppsStore.apps.length}</span>
				{/if}
			</button>
			<button
				class="btn-pill btn-pill-sm {activeTab === 'custom' ? 'btn-pill-primary' : 'btn-pill-ghost'}"
				onclick={() => (activeTab = 'custom')}
			>
				<Plus size={16} />
				Add Custom
			</button>
		</div>

		<!-- Content -->
		<div class="modal-content">
			{#if activeTab === 'browse'}
				<div class="browse-layout">
					<!-- Sidebar -->
					<aside class="sidebar">
						<div class="sidebar-section">
							<h4 class="sidebar-title">Categories</h4>
							<nav class="category-nav">
								{#each categories as cat}
									<button
										class="category-btn"
										class:active={selectedCategory === cat.id}
										onclick={() => (selectedCategory = cat.id)}
									>
										<svelte:component this={cat.icon} size={16} />
										<span class="category-name">{cat.name}</span>
										<span class="category-count">{getCategoryCount(cat.id)}</span>
									</button>
								{/each}
							</nav>
						</div>

						<!-- Your Apps Count -->
						{#if $userAppsStore.apps.length > 0}
							<div class="sidebar-section your-apps-section">
								<h4 class="sidebar-title">Your Apps</h4>
								<div class="your-apps-count">
									<Check size={16} />
									<span>{$userAppsStore.apps.length} apps added</span>
								</div>
							</div>
						{/if}
					</aside>

					<!-- Main Content -->
					<main class="main-content">
						<!-- Search -->
						<div class="search-wrapper">
							<Search size={18} class="search-icon" />
							<input
								type="text"
								bind:value={searchQuery}
								placeholder="Search {selectedCategory === 'all' ? 'all' : categories.find((c) => c.id === selectedCategory)?.name} apps..."
								class="search-input"
							/>
						</div>

						<div class="content-scroll">
							<!-- Featured Section (only on All) -->
							{#if selectedCategory === 'all' && !searchQuery}
								<section class="featured-section">
									<div class="section-header">
										<Star size={18} class="section-icon" />
										<h3>Featured Apps</h3>
									</div>
									<div class="featured-grid">
										{#each featuredApps as app (app.name)}
											<button
												class="featured-card"
												class:added={addedAppUrls.includes(app.url)}
												onclick={() => handleAppClick(app)}
												disabled={isSubmitting}
												transition:scale={{ duration: 200 }}
											>
												<div class="featured-logo">
													<img src={app.logo} alt={app.name} />
												</div>
												<div class="featured-info">
													<span class="featured-name">{app.name}</span>
													<span class="featured-desc">{app.description}</span>
												</div>
												<div class="featured-action">
													{#if addedAppUrls.includes(app.url)}
														<span class="added-icon"><Check size={16} /></span>
														<span class="remove-icon"><Trash2 size={16} /></span>
													{:else}
														<Plus size={16} />
													{/if}
												</div>
											</button>
										{/each}
									</div>
								</section>
							{/if}

							<!-- All Apps Grid -->
							<section class="apps-section">
								<div class="section-header">
									<Grid3X3 size={18} class="section-icon" />
									<h3>
										{selectedCategory === 'all'
											? 'All Apps'
											: categories.find((c) => c.id === selectedCategory)?.name}
									</h3>
									<span class="app-count">{filteredApps.length} apps</span>
								</div>

								{#each groupedApps() as group}
									{#if group.subcategory}
										<div class="subcategory-header">
											<span class="subcategory-name">{group.subcategory}</span>
											<span class="subcategory-count">{group.apps.length}</span>
										</div>
									{/if}
									<div class="apps-grid">
										{#each group.apps as app (app.name)}
											<button
												class="app-card"
												class:added={addedAppUrls.includes(app.url)}
												onclick={() => handleAppClick(app)}
												disabled={isSubmitting}
											>
												<div class="app-logo">
													<img src={app.logo} alt={app.name} />
												</div>
												<div class="app-info">
													<span class="app-name">{app.name}</span>
													<span class="app-desc">{app.description}</span>
												</div>
												<div class="app-action">
													{#if addedAppUrls.includes(app.url)}
														<div class="added-badge">
															<Check size={14} />
														</div>
														<div class="remove-badge">
															<Trash2 size={14} />
														</div>
													{:else}
														<div class="add-btn">
															<Plus size={14} />
														</div>
													{/if}
												</div>
											</button>
										{/each}
									</div>
								{/each}

								{#if filteredApps.length === 0}
									<div class="empty-state">
										<Search size={40} strokeWidth={1.5} />
										<p>No apps found matching "{searchQuery}"</p>
										<button class="btn-pill btn-pill-secondary btn-pill-sm clear-search" onclick={() => (searchQuery = '')}>
											Clear search
										</button>
									</div>
								{/if}
							</section>
						</div>
					</main>
				</div>
			{:else if activeTab === 'myapps'}
				<!-- My Apps -->
				<div class="myapps-wrapper">
					{#if $userAppsStore.apps.length === 0}
						<div class="myapps-empty">
							<Layers size={48} strokeWidth={1.5} />
							<h3>No apps installed yet</h3>
							<p>Browse the App Store and add apps to your desktop</p>
							<button class="btn-pill btn-pill-primary browse-btn" onclick={() => (activeTab = 'browse')}>
								<Globe size={18} />
								Browse Apps
							</button>
						</div>
					{:else}
						<div class="myapps-header">
							<h3>Installed Apps</h3>
							<p>{$userAppsStore.apps.length} apps on your desktop</p>
						</div>
						<div class="myapps-list">
							{#each $userAppsStore.apps as app (app.id)}
								<div class="myapp-item" class:disabled={app.is_active === false}>
									<div class="myapp-logo">
										{#if app.logo_url}
											<img src={app.logo_url} alt={app.name} />
										{:else}
											<div class="myapp-logo-placeholder" style="background: {app.color}">
												{app.name.charAt(0)}
											</div>
										{/if}
									</div>
									<div class="myapp-info">
										<span class="myapp-name">{app.name}</span>
										{#if app.description}
											<span class="myapp-desc">{app.description}</span>
										{:else}
											<span class="myapp-url">{app.url}</span>
										{/if}
									</div>
									<div class="myapp-actions">
										<button
											class="myapp-toggle"
											class:active={app.is_active !== false}
											onclick={() => toggleAppActive(app)}
											title={app.is_active !== false ? 'Disable app' : 'Enable app'}
										>
											{#if app.is_active !== false}
												<ToggleRight size={24} />
											{:else}
												<ToggleLeft size={24} />
											{/if}
										</button>
										<button
											class="myapp-remove"
											onclick={() => removeApp(app.id)}
											title="Remove app"
										>
											<Trash2 size={18} />
										</button>
									</div>
								</div>
							{/each}
						</div>
					{/if}
				</div>
			{:else}
				<!-- Custom App - Clean centered form -->
				<div class="custom-page">
					<div class="custom-form-container">
						<div class="custom-form-header">
							<h2 class="custom-form-title">Add Custom App</h2>
							<p class="custom-form-subtitle">Add any web application to your workspace</p>
						</div>

						<form onsubmit={handleCustomSubmit} class="custom-form">
							<div class="custom-field">
								<label for="app-name" class="custom-label">App Name</label>
								<input
									id="app-name"
									type="text"
									bind:value={name}
									placeholder="e.g., My Custom Tool"
									required
									class="custom-input"
								/>
							</div>

							<div class="custom-field">
								<label for="app-url" class="custom-label">URL</label>
								<input
									id="app-url"
									type="url"
									bind:value={url}
									placeholder="https://example.com"
									required
									class="custom-input"
								/>
							</div>

							<div class="custom-field-row">
								<div class="custom-field">
									<label for="app-category" class="custom-label">Category</label>
									<select id="app-category" bind:value={category} class="custom-select">
										{#each categories.filter((c) => c.id !== 'all') as cat}
											<option value={cat.id}>{cat.name}</option>
										{/each}
										<option value="other">Other</option>
									</select>
								</div>

								<div class="custom-field">
									<label for="app-description" class="custom-label">Description</label>
									<input
										id="app-description"
										type="text"
										bind:value={description}
										placeholder="Optional"
										class="custom-input"
									/>
								</div>
							</div>

							<!-- Live Preview -->
							{#if name || url}
								<div class="custom-preview">
									<span class="custom-preview-label">Preview</span>
									<div class="custom-preview-card">
										<div class="custom-preview-icon">
											{#if url}
												<img src="https://www.google.com/s2/favicons?domain={url.replace('https://', '').replace('http://', '').split('/')[0]}&sz=128" alt="" />
											{:else}
												<Globe size={20} />
											{/if}
										</div>
										<div class="custom-preview-info">
											<span class="custom-preview-name">{name || 'App Name'}</span>
											<span class="custom-preview-desc">{description || url || 'Description'}</span>
										</div>
									</div>
								</div>
							{/if}

							<button type="submit" class="custom-submit-btn" disabled={!name || !url || isSubmitting}>
								{#if isSubmitting}
									Adding...
								{:else}
									Add to My Apps
								{/if}
							</button>
						</form>
					</div>
				</div>
			{/if}
		</div>
	</div>
</div>

<style>
	/* Page mode styles */
	.page-container {
		height: 100%;
		width: 100%;
		display: flex;
		flex-direction: column;
		background: #f8fafc;
	}

	.page-content {
		flex: 1;
		display: flex;
		flex-direction: column;
		overflow: hidden;
		background: #f8fafc;
	}

	.page-header {
		padding: 20px 24px;
		border-bottom: 1px solid #e2e8f0;
		background: white;
	}

	.page-header .header-text h2 {
		font-size: 1.5rem;
		font-weight: 600;
		color: #111827;
		margin: 0;
	}

	.page-header .header-text p {
		font-size: 0.875rem;
		color: #6b7280;
		margin: 4px 0 0;
	}

	/* Make modal-content work in page mode */
	.page-container .modal-content {
		flex: 1;
		overflow: hidden;
	}

	.page-container .main-tabs {
		background: white;
		padding: 0.875rem 1.5rem;
		border-bottom: 1px solid #e5e7eb;
		gap: 0.75rem;
	}

	.page-container .main-tab {
		color: #6b7280;
		background: #f3f4f6;
		border: 1px solid #e5e7eb;
		padding: 0.625rem 1.25rem;
		border-radius: 0.5rem;
		font-weight: 500;
	}

	.page-container .main-tab:hover {
		background: #e5e7eb;
		color: #374151;
		border-color: #d1d5db;
	}

	.page-container .main-tab.active {
		background: #0A84FF !important;
		color: white !important;
		border-color: #0A84FF !important;
	}

	/* Page mode - Light sidebar */
	.page-container .sidebar {
		background: #f9fafb;
		border-right-color: #e5e7eb;
	}

	.page-container .sidebar-title {
		color: #6b7280;
	}

	.page-container .category-btn {
		color: #6b7280;
		background: transparent;
	}

	.page-container .category-btn:hover {
		background: #e5e7eb;
		color: #374151;
	}

	.page-container .category-btn.active {
		background: #0A84FF !important;
		color: white !important;
	}

	.page-container .category-btn.active :global(svg) {
		color: white !important;
	}

	.page-container .category-count {
		color: #6b7280;
	}

	/* Page mode - Light content area */
	.page-container .main-content {
		background: #f8fafc;
	}

	.page-container .search-wrapper {
		background: white;
		border-color: #e5e7eb;
	}

	.page-container .search-input {
		background: transparent;
		color: #111827;
	}

	.page-container .search-input::placeholder {
		color: #9ca3af;
	}

	/* Page mode - Featured Apps section */
	.page-container .featured-section {
		background: white !important;
		border: 1px solid #e5e7eb !important;
		border-radius: 16px !important;
		padding: 1.25rem !important;
		margin: 0 0 1.5rem 0 !important;
	}

	.page-container .section-header {
		margin-bottom: 1rem !important;
	}

	.page-container .section-header h3 {
		color: #111827;
		font-size: 1rem;
	}

	.page-container .section-icon {
		color: #0A84FF !important;
	}

	.page-container .featured-grid {
		background: transparent;
		gap: 12px !important;
		padding: 4px !important;
	}

	.page-container .featured-card {
		background: #f9fafb !important;
		border: 1px solid #e5e7eb !important;
		border-radius: 14px !important;
		padding: 14px !important;
		transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1) !important;
		gap: 12px !important;
	}

	.page-container .featured-card:hover:not(:disabled) {
		background: white !important;
		border-color: #d1d5db !important;
		transform: translateY(-1px) !important;
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08) !important;
	}

	.page-container .featured-card.added {
		background: #f0fdf4 !important;
		border-color: #22c55e !important;
	}

	.page-container .featured-card.added:hover {
		background: #f0fdf4 !important;
		border-color: #16a34a !important;
	}

	.page-container .featured-name {
		color: #111827 !important;
		font-weight: 500 !important;
		white-space: normal !important;
		overflow: visible !important;
	}

	.page-container .featured-desc {
		color: #6b7280 !important;
		font-size: 0.75rem !important;
		white-space: normal !important;
		overflow: visible !important;
	}

	.page-container .featured-info {
		flex: 1 !important;
		min-width: 0 !important;
		overflow: visible !important;
	}

	.page-container .featured-logo {
		border-radius: 10px !important;
		width: 36px !important;
		height: 36px !important;
	}

	.page-container .featured-logo img {
		border-radius: 8px !important;
	}

	/* Featured card action button - subtle by default */
	.page-container .featured-action {
		background: transparent !important;
		color: #9ca3af !important;
		border-radius: 8px !important;
		width: 32px !important;
		height: 32px !important;
		min-width: 32px !important;
		transition: all 0.2s ease !important;
		display: flex !important;
		align-items: center !important;
		justify-content: center !important;
		border: 1px solid #d1d5db !important;
	}

	/* Hover on non-added card - show blue */
	.page-container .featured-card:hover:not(:disabled):not(.added) .featured-action {
		background: #0A84FF !important;
		color: white !important;
		border-color: #0A84FF !important;
	}

	/* Added card - show checkmark with green */
	.page-container .featured-card.added .featured-action {
		background: #22c55e !important;
		color: white !important;
		border: none !important;
	}

	/* Hover on added card - red for remove */
	.page-container .featured-card.added:hover .featured-action {
		background: #ef4444 !important;
		color: white !important;
		border: none !important;
	}

	/* Featured card red hover for remove action */
	.page-container .featured-card.added:hover {
		background: #fef2f2 !important;
		border-color: #ef4444 !important;
	}

	/* Featured card icon visibility in page mode */
	.page-container .featured-card .added-icon {
		display: block !important;
	}

	.page-container .featured-card .remove-icon {
		display: none !important;
	}

	.page-container .featured-card.added:hover .added-icon {
		display: none !important;
	}

	.page-container .featured-card.added:hover .remove-icon {
		display: block !important;
		color: white !important;
	}

	/* Page mode - App card action container - NO background, children handle it */
	.page-container .app-action {
		background: transparent !important;
		display: flex !important;
		align-items: center !important;
		justify-content: center !important;
	}

	.page-container .apps-header h3 {
		color: #111827;
	}

	.page-container .apps-header span {
		color: #6b7280;
	}

	/* Page mode - Light app cards */
	.page-container .app-card {
		background: white !important;
		border: 1px solid #e5e7eb !important;
		border-radius: 14px !important;
		padding: 14px !important;
		transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1) !important;
		gap: 12px !important;
	}

	.page-container .app-card:hover:not(:disabled) {
		background: #f9fafb !important;
		border-color: #d1d5db !important;
		transform: translateY(-1px) !important;
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08) !important;
	}

	.page-container .app-card.added {
		background: #f0fdf4 !important;
		border-color: #22c55e !important;
	}

	.page-container .app-card.added:hover {
		background: #f0fdf4 !important;
		border-color: #16a34a !important;
	}

	.page-container .app-logo {
		border-radius: 10px !important;
		width: 36px !important;
		height: 36px !important;
	}

	.page-container .app-logo img {
		border-radius: 8px !important;
	}

	.page-container .app-name {
		color: #111827 !important;
		font-weight: 500 !important;
		white-space: normal !important;
		overflow: visible !important;
		text-overflow: unset !important;
	}

	.page-container .app-desc {
		color: #6b7280 !important;
		font-size: 0.75rem !important;
		white-space: normal !important;
		overflow: visible !important;
		text-overflow: unset !important;
	}

	.page-container .app-info {
		flex: 1 !important;
		min-width: 0 !important;
		overflow: visible !important;
	}

	.page-container .app-category {
		color: #6b7280;
	}

	/* Page mode - Subcategory headers */
	.page-container .subcategory-header {
		border-color: #e5e7eb !important;
		margin: 1.75rem 0 1.25rem 0 !important;
		padding-bottom: 0.625rem !important;
		border-bottom-width: 1px !important;
	}

	.page-container .subcategory-header:first-of-type {
		margin-top: 0.5rem !important;
	}

	.page-container .subcategory-name {
		color: #374151 !important;
		font-size: 0.8125rem !important;
		font-weight: 600 !important;
		letter-spacing: 0.02em !important;
	}

	.page-container .subcategory-count {
		color: #6b7280 !important;
		font-size: 0.75rem !important;
		background: #e5e7eb !important;
		padding: 0.125rem 0.5rem !important;
		border-radius: 10px !important;
	}

	/* Page mode - Content scroll area */
	.page-container .content-scroll {
		background: #f8fafc;
	}

	.page-container .app-count {
		color: #6b7280;
	}

	/* Page mode - Add button - subtle by default, blue on hover */
	.page-container .add-btn {
		background: transparent !important;
		color: #9ca3af !important;
		border-radius: 8px !important;
		width: 32px !important;
		height: 32px !important;
		min-width: 32px !important;
		transition: all 0.2s ease !important;
		display: flex !important;
		align-items: center !important;
		justify-content: center !important;
		border: 1px solid #d1d5db !important;
	}

	.page-container .app-card:hover:not(:disabled):not(.added) .add-btn {
		background: #0A84FF !important;
		color: white !important;
		border-color: #0A84FF !important;
	}

	/* Page mode - Added badge styling (checkmark) */
	.page-container .added-badge {
		background: #22c55e !important;
		color: white !important;
		border-radius: 8px !important;
		width: 32px !important;
		height: 32px !important;
		display: flex !important;
		align-items: center !important;
		justify-content: center !important;
	}

	/* Page mode - Remove badge styling (trash on hover) */
	.page-container .remove-badge {
		background: #ef4444 !important;
		color: white !important;
		border-radius: 8px !important;
		width: 32px !important;
		height: 32px !important;
		display: none !important;
		align-items: center !important;
		justify-content: center !important;
	}

	.page-container .app-card.added:hover .added-badge {
		display: none !important;
	}

	.page-container .app-card.added:hover .remove-badge {
		display: flex !important;
	}

	/* Added app card red hover for remove action */
	.page-container .app-card.added:hover {
		background: #fef2f2 !important;
		border-color: #ef4444 !important;
	}

	/* Page mode - Tab badge */
	.page-container .tab-badge {
		background: #0A84FF !important;
		color: white !important;
	}

	/* Page mode - Search focus */
	.page-container .search-wrapper:focus-within {
		border-color: #0A84FF;
	}

	/* Page mode - Apps grid scroll area */
	.page-container .apps-scroll {
		background: #f8fafc;
	}

	.page-container .apps-section {
		background: transparent;
		padding: 0 0.5rem !important;
	}

	.page-container .apps-grid {
		gap: 14px !important;
		padding: 8px 4px !important;
	}

	.page-container .featured-grid {
		gap: 14px !important;
		padding: 8px 4px !important;
	}

	.page-container .apps-section .section-header {
		margin-bottom: 1.25rem !important;
		padding-bottom: 0.75rem !important;
		border-bottom: 1px solid #e5e7eb !important;
	}

	.page-container .apps-section .section-header h3 {
		color: #111827;
		font-size: 1rem;
	}

	.page-container .section-count,
	.page-container .app-count {
		color: #6b7280 !important;
		font-size: 0.8125rem !important;
	}

	/* Page mode - Your apps section */
	.page-container .your-apps-count {
		background: #f3f4f6 !important;
		color: #6b7280 !important;
		padding: 0.625rem 0.875rem !important;
		border-radius: 10px !important;
		border: 1px solid #e5e7eb !important;
		font-size: 0.8125rem !important;
	}

	.page-container .your-apps-count :global(svg) {
		color: #0A84FF !important;
	}

	.page-container .your-apps-section {
		background: transparent !important;
		border-radius: 0.5rem;
		padding: 0.5rem 0.75rem !important;
		margin: 0 0.5rem;
		border-top-color: #e5e7eb !important;
	}

	.page-container .your-apps-section .sidebar-title {
		color: #6b7280;
		margin-bottom: 0.5rem;
	}

	/* Page mode - My Apps tab */
	.page-container .myapps-wrapper {
		background: white;
	}

	.page-container .myapps-header h3 {
		color: #111827;
	}

	.page-container .myapps-header p {
		color: #6b7280;
	}

	.page-container .myapps-grid {
		background: transparent;
	}

	.page-container .myapp-card {
		background: white;
		border-color: #e5e7eb;
	}

	.page-container .myapp-card:hover {
		background: #f9fafb;
		border-color: #d1d5db;
	}

	.page-container .myapp-name {
		color: #111827;
	}

	.page-container .myapp-url {
		color: #6b7280;
	}

	.page-container .myapp-category {
		color: #6b7280;
	}

	.page-container .myapp-actions button {
		background: #f3f4f6;
		color: #6b7280;
		border-color: #e5e7eb;
	}

	.page-container .myapp-actions button:hover {
		background: #e5e7eb;
		color: #374151;
	}

	.page-container .myapp-actions .myapp-delete:hover {
		background: rgba(239, 68, 68, 0.2);
		color: #ef4444;
		border-color: #ef4444;
	}

	.page-container .myapp-toggle.active {
		color: #0A84FF;
	}

	.page-container .empty-state {
		color: #6b7280;
	}

	.page-container .empty-state h3 {
		color: #fff;
	}

	.page-container .empty-cta {
		background: linear-gradient(135deg, #0A84FF 0%, #0066CC 100%);
	}

	.page-container .empty-cta:hover {
		background: linear-gradient(135deg, #0077ED 0%, #0055AA 100%);
	}

	/* Page mode - Custom form - CLEAN CENTERED */
	.page-container .custom-page {
		flex: 1 !important;
		display: flex !important;
		align-items: center !important;
		justify-content: center !important;
		padding: 2rem !important;
		background: #1a1a1a !important;
	}

	.page-container .custom-form-container {
		width: 100% !important;
		max-width: 480px !important;
		background: #222 !important;
		border-radius: 16px !important;
		padding: 2rem !important;
		border: 1px solid #333 !important;
	}

	.page-container .custom-form-header {
		text-align: center !important;
		margin-bottom: 2rem !important;
	}

	.page-container .custom-form-title {
		color: #fff !important;
		font-size: 1.5rem !important;
		font-weight: 600 !important;
		margin: 0 0 0.5rem 0 !important;
	}

	.page-container .custom-form-subtitle {
		color: #888 !important;
		font-size: 0.875rem !important;
		margin: 0 !important;
	}

	.page-container .custom-form {
		display: flex !important;
		flex-direction: column !important;
		gap: 1.25rem !important;
	}

	.page-container .custom-field {
		display: flex !important;
		flex-direction: column !important;
		gap: 0.5rem !important;
	}

	.page-container .custom-field-row {
		display: grid !important;
		grid-template-columns: 1fr 1fr !important;
		gap: 1rem !important;
	}

	.page-container .custom-label {
		color: #999 !important;
		font-size: 0.8125rem !important;
		font-weight: 500 !important;
	}

	.page-container .custom-input,
	.page-container .custom-select {
		background: #1a1a1a !important;
		border: 1px solid #3a3a3a !important;
		border-radius: 10px !important;
		padding: 0.875rem 1rem !important;
		color: #fff !important;
		font-size: 0.9375rem !important;
		transition: all 0.2s ease !important;
		width: 100% !important;
	}

	.page-container .custom-input:focus,
	.page-container .custom-select:focus {
		border-color: #0A84FF !important;
		outline: none !important;
		box-shadow: 0 0 0 3px rgba(10, 132, 255, 0.15) !important;
	}

	.page-container .custom-input::placeholder {
		color: #555 !important;
	}

	.page-container .custom-preview {
		background: #1a1a1a !important;
		border-radius: 10px !important;
		padding: 1rem !important;
		margin-top: 0.5rem !important;
	}

	.page-container .custom-preview-label {
		color: #666 !important;
		font-size: 0.6875rem !important;
		font-weight: 600 !important;
		text-transform: uppercase !important;
		letter-spacing: 0.05em !important;
		display: block !important;
		margin-bottom: 0.75rem !important;
	}

	.page-container .custom-preview-card {
		display: flex !important;
		align-items: center !important;
		gap: 0.75rem !important;
		background: #252525 !important;
		border-radius: 10px !important;
		padding: 0.75rem !important;
	}

	.page-container .custom-preview-icon {
		width: 36px !important;
		height: 36px !important;
		border-radius: 8px !important;
		background: #333 !important;
		display: flex !important;
		align-items: center !important;
		justify-content: center !important;
		flex-shrink: 0 !important;
	}

	.page-container .custom-preview-icon img {
		width: 20px !important;
		height: 20px !important;
		border-radius: 4px !important;
	}

	.page-container .custom-preview-icon :global(svg) {
		color: #888 !important;
	}

	.page-container .custom-preview-info {
		display: flex !important;
		flex-direction: column !important;
		gap: 0.125rem !important;
		min-width: 0 !important;
	}

	.page-container .custom-preview-name {
		color: #fff !important;
		font-size: 0.875rem !important;
		font-weight: 500 !important;
	}

	.page-container .custom-preview-desc {
		color: #666 !important;
		font-size: 0.75rem !important;
		white-space: nowrap !important;
		overflow: hidden !important;
		text-overflow: ellipsis !important;
	}

	.page-container .custom-submit-btn {
		background: #0A84FF !important;
		color: white !important;
		border: none !important;
		border-radius: 10px !important;
		padding: 1rem !important;
		font-size: 0.9375rem !important;
		font-weight: 600 !important;
		cursor: pointer !important;
		transition: all 0.2s ease !important;
		margin-top: 0.5rem !important;
	}

	.page-container .custom-submit-btn:hover:not(:disabled) {
		background: #0077ED !important;
	}

	.page-container .custom-submit-btn:disabled {
		background: #333 !important;
		color: #555 !important;
		cursor: not-allowed !important;
		margin-bottom: 1rem !important;
	}

	.page-container .preview-card {
		background: #252525 !important;
		border: 1px solid #3a3a3a !important;
		border-radius: 12px !important;
		padding: 1rem !important;
		display: flex !important;
		align-items: center !important;
		gap: 1rem !important;
		max-width: 350px !important;
	}

	.page-container .preview-logo {
		width: 40px !important;
		height: 40px !important;
		border-radius: 10px !important;
		background: #333 !important;
		display: flex !important;
		align-items: center !important;
		justify-content: center !important;
		overflow: hidden !important;
		color: #666 !important;
	}

	.page-container .preview-logo img {
		width: 100% !important;
		height: 100% !important;
		object-fit: cover !important;
	}

	.page-container .preview-info {
		flex: 1 !important;
		display: flex !important;
		flex-direction: column !important;
		gap: 0.25rem !important;
	}

	.page-container .preview-name {
		color: #fff !important;
		font-weight: 500 !important;
		font-size: 0.9375rem !important;
	}

	.page-container .preview-desc {
		color: #666 !important;
		font-size: 0.8125rem !important;
	}

	.page-container .preview-action {
		width: 28px !important;
		height: 28px !important;
		border-radius: 8px !important;
		background: #22c55e !important;
		color: white !important;
		display: flex !important;
		align-items: center !important;
		justify-content: center !important;
	}

	/* Tips section in sidebar */
	.page-container .tips-section {
		margin-top: 2rem !important;
	}

	.page-container .tips-list {
		display: flex !important;
		flex-direction: column !important;
		gap: 0.75rem !important;
	}

	.page-container .tip-item {
		display: flex !important;
		align-items: center !important;
		gap: 0.625rem !important;
		color: #666 !important;
		font-size: 0.8125rem !important;
	}

	.page-container .tip-item :global(svg) {
		color: #0A84FF !important;
		flex-shrink: 0 !important;
	}

	.page-container .sidebar-desc {
		color: #666 !important;
		font-size: 0.8125rem !important;
		margin-top: 0.5rem !important;
	}

	.page-container .form-header h3 {
		color: #ffffff !important;
		font-size: 1.375rem !important;
		font-weight: 600 !important;
		margin-top: 1.25rem !important;
		margin-bottom: 0.5rem !important;
	}

	.page-container .form-header p {
		color: #888 !important;
		font-size: 0.875rem !important;
	}

	.page-container .form-group {
		margin-bottom: 1.5rem !important;
	}

	.page-container .form-group label,
	.page-container .form-label {
		color: #ccc !important;
		font-weight: 500 !important;
		font-size: 0.8125rem !important;
		margin-bottom: 0.5rem !important;
		display: block !important;
	}

	.page-container .form-group .required {
		color: #0A84FF !important;
	}

	.page-container .form-input,
	.page-container .form-select,
	.page-container .form-textarea,
	.page-container .form-group input,
	.page-container .form-group select,
	.page-container .form-group textarea {
		background: #0d0d0d !important;
		border: 1px solid #333 !important;
		color: #fff !important;
		border-radius: 12px !important;
		padding: 1rem 1.125rem !important;
		font-size: 0.9375rem !important;
		width: 100% !important;
		transition: all 0.2s ease !important;
	}

	.page-container .form-group input::placeholder,
	.page-container .form-group textarea::placeholder,
	.page-container .form-input::placeholder,
	.page-container .form-textarea::placeholder {
		color: #555 !important;
	}

	.page-container .form-group input:focus,
	.page-container .form-group select:focus,
	.page-container .form-group textarea:focus,
	.page-container .form-input:focus,
	.page-container .form-select:focus,
	.page-container .form-textarea:focus {
		border-color: #0A84FF !important;
		box-shadow: 0 0 0 3px rgba(10, 132, 255, 0.2) !important;
		outline: none !important;
		background: #111 !important;
	}

	.page-container .url-input-wrapper {
		background: transparent !important;
		border: none !important;
		position: relative !important;
	}

	.page-container .url-input-wrapper .form-input,
	.page-container .url-input-wrapper input {
		padding-left: 2.75rem !important;
	}

	.page-container .url-icon,
	.page-container .url-input-wrapper :global(svg) {
		color: #666 !important;
		position: absolute !important;
		left: 1rem !important;
		top: 50% !important;
		transform: translateY(-50%) !important;
	}

	.page-container .input-hint {
		color: #666 !important;
		font-size: 0.75rem !important;
		margin-top: 0.5rem !important;
	}

	.page-container .submit-btn {
		background: linear-gradient(135deg, #0A84FF 0%, #0066CC 100%) !important;
		border: none !important;
		padding: 1.125rem 1.5rem !important;
		font-size: 1rem !important;
		font-weight: 600 !important;
		border-radius: 12px !important;
		box-shadow: 0 8px 24px rgba(10, 132, 255, 0.35) !important;
		color: white !important;
		width: 100% !important;
		display: flex !important;
		align-items: center !important;
		justify-content: center !important;
		gap: 0.5rem !important;
		margin-top: 2.5rem !important;
		transition: all 0.25s ease !important;
		cursor: pointer !important;
	}

	.page-container .submit-btn:hover:not(:disabled) {
		background: linear-gradient(135deg, #0077ED 0%, #0055AA 100%) !important;
		transform: translateY(-2px) !important;
		box-shadow: 0 12px 32px rgba(10, 132, 255, 0.45) !important;
	}

	.page-container .submit-btn:active:not(:disabled) {
		transform: translateY(0) !important;
	}

	.page-container .submit-btn:disabled {
		background: #252525 !important;
		color: #555 !important;
		box-shadow: none !important;
		transform: none !important;
		cursor: not-allowed !important;
	}

	.modal-overlay {
		position: fixed;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background: rgba(0, 0, 0, 0.6);
		backdrop-filter: blur(8px);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 10000;
		padding: 1rem;
	}

	.modal-container {
		background: #fafafa;
		border-radius: 1rem;
		box-shadow:
			0 25px 50px -12px rgba(0, 0, 0, 0.25),
			0 0 0 1px rgba(0, 0, 0, 0.05);
		width: 100%;
		max-width: 1100px;
		height: 85vh;
		max-height: 800px;
		display: flex;
		flex-direction: column;
		overflow: hidden;
	}

	/* Header */
	.modal-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 1.25rem 1.5rem;
		background: white;
		border-bottom: 1px solid #e5e7eb;
	}

	.header-content {
		display: flex;
		align-items: center;
		gap: 0.875rem;
	}

	.header-text h2 {
		font-size: 1.25rem;
		font-weight: 700;
		color: #111827;
		margin: 0 0 0.125rem 0;
	}

	.header-text p {
		font-size: 0.8125rem;
		color: #6b7280;
		margin: 0;
	}

	.close-btn {
		background: #f3f4f6;
		border: none;
		border-radius: 0.5rem;
		width: 34px;
		height: 34px;
		display: flex;
		align-items: center;
		justify-content: center;
		cursor: pointer;
		color: #6b7280;
		transition: all 0.2s;
	}

	.close-btn:hover {
		background: #e5e7eb;
		color: #111827;
	}

	/* Main Tabs */
	.main-tabs {
		display: flex;
		gap: 0.25rem;
		padding: 0.5rem 1.5rem;
		background: white;
		border-bottom: 1px solid #e5e7eb;
	}

	.main-tab {
		background: none;
		border: none;
		padding: 0.625rem 1rem;
		font-size: 0.875rem;
		font-weight: 500;
		color: #6b7280;
		cursor: pointer;
		display: flex;
		align-items: center;
		gap: 0.5rem;
		border-radius: 0.5rem;
		transition: all 0.2s;
	}

	.main-tab:hover {
		background: #f3f4f6;
		color: #111827;
	}

	.main-tab.active {
		background: #10B981;
		color: white;
	}

	/* Content */
	.modal-content {
		flex: 1;
		overflow: hidden;
	}

	/* Browse Layout */
	.browse-layout {
		display: flex;
		height: 100%;
	}

	/* Sidebar */
	.sidebar {
		width: 220px;
		background: white;
		border-right: 1px solid #e5e7eb;
		padding: 1rem 0;
		overflow-y: auto;
		flex-shrink: 0;
	}

	.sidebar-section {
		padding: 0 0.75rem;
		margin-bottom: 1.5rem;
	}

	.sidebar-title {
		font-size: 0.6875rem;
		font-weight: 600;
		color: #9ca3af;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		padding: 0 0.5rem;
		margin: 0 0 0.5rem 0;
	}

	.category-nav {
		display: flex;
		flex-direction: column;
		gap: 0.125rem;
	}

	.category-btn {
		display: flex;
		align-items: center;
		gap: 0.625rem;
		padding: 0.5rem 0.625rem;
		background: none;
		border: none;
		border-radius: 0.5rem;
		cursor: pointer;
		transition: all 0.15s;
		text-align: left;
		color: #4b5563;
	}

	.category-btn:hover {
		background: #f3f4f6;
		color: #111827;
	}

	.category-btn.active {
		background: #ecfdf5;
		color: #10B981;
	}

	.category-btn.active :global(svg) {
		color: #10B981;
	}

	.category-name {
		flex: 1;
		font-size: 0.8125rem;
		font-weight: 500;
	}

	.category-count {
		font-size: 0.75rem;
		color: #9ca3af;
		background: #f3f4f6;
		padding: 0.125rem 0.375rem;
		border-radius: 0.25rem;
	}

	.category-btn.active .category-count {
		background: #a7f3d0;
		color: #047857;
	}

	.your-apps-section {
		border-top: 1px solid #e5e7eb;
		padding-top: 1rem;
	}

	.your-apps-count {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.5rem 0.625rem;
		font-size: 0.8125rem;
		color: #059669;
		background: #ecfdf5;
		border-radius: 0.5rem;
	}

	/* Main Content */
	.main-content {
		flex: 1;
		display: flex;
		flex-direction: column;
		overflow: hidden;
	}

	.search-wrapper {
		position: relative;
		padding: 1rem 1.5rem;
		background: white;
		border-bottom: 1px solid #f3f4f6;
	}

	:global(.search-icon) {
		position: absolute;
		left: 2.25rem;
		top: 50%;
		transform: translateY(-50%);
		color: #9ca3af;
		pointer-events: none;
	}

	.search-input {
		width: 100%;
		padding: 0.625rem 1rem 0.625rem 2.5rem;
		border: 1px solid #e5e7eb;
		border-radius: 0.5rem;
		font-size: 0.875rem;
		transition: all 0.2s;
		background: #f9fafb;
	}

	.search-input:focus {
		outline: none;
		border-color: #10B981;
		background: white;
		box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.1);
	}

	.content-scroll {
		flex: 1;
		overflow-y: auto;
		padding: 1.25rem 1.5rem;
	}

	/* Section Headers */
	.section-header {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		margin-bottom: 1rem;
	}

	:global(.section-icon) {
		color: #6b7280;
	}

	.section-header h3 {
		font-size: 0.9375rem;
		font-weight: 600;
		color: #111827;
		margin: 0;
	}

	.app-count {
		font-size: 0.75rem;
		color: #9ca3af;
		margin-left: auto;
	}

	/* Featured Section */
	.featured-section {
		margin-bottom: 2rem;
	}

	.featured-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
		gap: 0.75rem;
	}

	.featured-card {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		padding: 0.875rem;
		background: white;
		border: 1px solid #e5e7eb;
		border-radius: 0.75rem;
		cursor: pointer;
		transition: all 0.2s;
		text-align: left;
	}

	.featured-card:hover:not(:disabled) {
		border-color: #10B981;
		box-shadow: 0 4px 12px -2px rgba(99, 102, 241, 0.15);
		transform: translateY(-1px);
	}

	.featured-card.added {
		background: #f0fdf4;
		border-color: #86efac;
	}

	.featured-card:disabled {
		cursor: default;
	}

	.featured-logo {
		width: 40px;
		height: 40px;
		border-radius: 0.5rem;
		overflow: hidden;
		background: #f9fafb;
		flex-shrink: 0;
	}

	.featured-logo img {
		width: 100%;
		height: 100%;
		object-fit: contain;
	}

	.featured-info {
		flex: 1;
		min-width: 0;
	}

	.featured-name {
		display: block;
		font-size: 0.875rem;
		font-weight: 600;
		color: #111827;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.featured-desc {
		display: block;
		font-size: 0.75rem;
		color: #6b7280;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.featured-action {
		width: 28px;
		height: 28px;
		display: flex;
		align-items: center;
		justify-content: center;
		background: #f3f4f6;
		border-radius: 0.375rem;
		color: #6b7280;
		flex-shrink: 0;
	}

	.featured-card:hover:not(:disabled):not(.added) .featured-action {
		background: #10B981;
		color: white;
	}

	.featured-card.added .featured-action {
		background: #22c55e;
		color: white;
	}

	/* Apps Grid */
	.apps-section {
		margin-bottom: 1rem;
	}

	/* Subcategory Headers */
	.subcategory-header {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.75rem 0;
		margin-top: 1rem;
		border-bottom: 1px solid #e5e7eb;
	}

	.subcategory-header:first-child {
		margin-top: 0;
	}

	.subcategory-name {
		font-size: 0.8125rem;
		font-weight: 600;
		color: #374151;
		text-transform: uppercase;
		letter-spacing: 0.025em;
	}

	.subcategory-count {
		font-size: 0.6875rem;
		color: #9ca3af;
		background: #f3f4f6;
		padding: 0.125rem 0.375rem;
		border-radius: 0.25rem;
	}

	.apps-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
		gap: 0.625rem;
	}

	.app-card {
		display: flex;
		align-items: center;
		gap: 0.625rem;
		padding: 0.625rem 0.75rem;
		background: white;
		border: 1px solid #e5e7eb;
		border-radius: 0.5rem;
		cursor: pointer;
		transition: all 0.15s;
		text-align: left;
	}

	.app-card:hover:not(:disabled) {
		border-color: #a7f3d0;
		background: #fafafa;
	}

	.app-card.added {
		background: #f0fdf4;
		border-color: #bbf7d0;
	}

	.app-card:disabled {
		cursor: default;
	}

	.app-logo {
		width: 32px;
		height: 32px;
		border-radius: 0.375rem;
		overflow: hidden;
		background: #f9fafb;
		flex-shrink: 0;
	}

	.app-logo img {
		width: 100%;
		height: 100%;
		object-fit: contain;
	}

	.app-info {
		flex: 1;
		min-width: 0;
	}

	.app-name {
		display: block;
		font-size: 0.8125rem;
		font-weight: 500;
		color: #111827;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.app-desc {
		display: block;
		font-size: 0.6875rem;
		color: #9ca3af;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.app-action {
		flex-shrink: 0;
	}

	.add-btn {
		width: 24px;
		height: 24px;
		display: flex;
		align-items: center;
		justify-content: center;
		background: #f3f4f6;
		border-radius: 0.375rem;
		color: #6b7280;
		transition: all 0.15s;
	}

	.app-card:hover:not(:disabled):not(.added) .add-btn {
		background: #10B981;
		color: white;
	}

	.added-badge {
		width: 24px;
		height: 24px;
		display: flex;
		align-items: center;
		justify-content: center;
		background: #22c55e;
		border-radius: 0.375rem;
		color: white;
	}

	/* Empty State */
	.empty-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 3rem 1rem;
		text-align: center;
		color: #9ca3af;
	}

	.empty-state p {
		margin: 1rem 0 0.5rem;
		font-size: 0.875rem;
	}

	.clear-search {
		background: #f3f4f6;
		border: none;
		padding: 0.5rem 1rem;
		border-radius: 0.375rem;
		font-size: 0.8125rem;
		color: #6b7280;
		cursor: pointer;
		transition: all 0.15s;
	}

	.clear-search:hover {
		background: #e5e7eb;
		color: #111827;
	}

	/* Custom Form */
	.custom-form-wrapper {
		display: flex;
		align-items: center;
		justify-content: center;
		height: 100%;
		padding: 2rem;
		background: white;
	}

	.custom-form {
		width: 100%;
		max-width: 420px;
	}

	.form-header {
		text-align: center;
		margin-bottom: 2rem;
		color: #6b7280;
	}

	.form-header h3 {
		font-size: 1.25rem;
		font-weight: 600;
		color: #111827;
		margin: 1rem 0 0.25rem;
	}

	.form-header p {
		font-size: 0.875rem;
		color: #6b7280;
		margin: 0;
	}

	.form-group {
		margin-bottom: 1.25rem;
	}

	.form-label {
		display: block;
		font-size: 0.8125rem;
		font-weight: 500;
		color: #374151;
		margin-bottom: 0.375rem;
	}

	.required {
		color: #0A84FF;
	}

	.form-input,
	.form-select,
	.form-textarea {
		width: 100%;
		padding: 0.625rem 0.875rem;
		border: 1px solid #e5e7eb;
		border-radius: 0.5rem;
		font-size: 0.875rem;
		transition: all 0.2s;
		background: white;
	}

	.form-input:focus,
	.form-select:focus,
	.form-textarea:focus {
		outline: none;
		border-color: #0A84FF;
		box-shadow: 0 0 0 3px rgba(10, 132, 255, 0.1);
	}

	.url-input-wrapper {
		position: relative;
	}

	.url-input {
		padding-left: 2.5rem;
	}

	:global(.url-icon) {
		position: absolute;
		left: 0.875rem;
		top: 50%;
		transform: translateY(-50%);
		color: #9ca3af;
		pointer-events: none;
	}

	.input-hint {
		display: block;
		font-size: 0.75rem;
		color: #9ca3af;
		margin-top: 0.375rem;
	}

	.form-textarea {
		resize: vertical;
		min-height: 60px;
	}

	.submit-btn {
		width: 100%;
		background: linear-gradient(135deg, #10B981 0%, #059669 100%);
		color: white;
		border: none;
		padding: 0.75rem 1.25rem;
		border-radius: 0.5rem;
		font-weight: 600;
		font-size: 0.875rem;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 0.5rem;
		transition: all 0.2s;
		margin-top: 1.5rem;
	}

	.submit-btn:hover:not(:disabled) {
		transform: translateY(-1px);
		box-shadow: 0 4px 12px -2px rgba(99, 102, 241, 0.4);
	}

	.submit-btn:disabled {
		opacity: 0.6;
		cursor: not-allowed;
		transform: none;
	}

	.spinner {
		width: 16px;
		height: 16px;
		border: 2px solid rgba(255, 255, 255, 0.3);
		border-top-color: white;
		border-radius: 50%;
		animation: spin 0.6s linear infinite;
	}

	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}

	/* ===== DARK MODE STYLES ===== */
	:global(.dark) .modal-overlay {
		background: rgba(0, 0, 0, 0.8);
	}

	:global(.dark) .modal-container {
		background: #1a1a1a;
		box-shadow:
			0 25px 50px -12px rgba(0, 0, 0, 0.5),
			0 0 0 1px rgba(255, 255, 255, 0.1);
	}

	:global(.dark) .modal-header {
		background: #1f1f1f;
		border-color: #333;
	}

	:global(.dark) .header-text h2 {
		color: #ffffff;
	}

	:global(.dark) .header-text p {
		color: #9ca3af;
	}

	:global(.dark) .close-btn {
		background: #333;
		color: #9ca3af;
	}

	:global(.dark) .close-btn:hover {
		background: #444;
		color: #ffffff;
	}

	:global(.dark) .main-tabs {
		background: #1f1f1f;
		border-color: #333;
	}

	:global(.dark) .main-tab {
		color: #9ca3af;
	}

	:global(.dark) .main-tab:hover {
		background: #333;
		color: #ffffff;
	}

	:global(.dark) .main-tab.active {
		background: #10B981;
		color: white;
	}

	:global(.dark) .sidebar {
		background: #1f1f1f;
		border-color: #333;
	}

	:global(.dark) .sidebar-title {
		color: #9ca3af;
	}

	:global(.dark) .category-btn {
		color: #d1d5db;
	}

	:global(.dark) .category-btn:hover {
		background: #333;
	}

	:global(.dark) .category-btn.active {
		background: rgba(99, 102, 241, 0.15);
		color: #34d399;
	}

	:global(.dark) .category-count {
		background: #333;
		color: #9ca3af;
	}

	:global(.dark) .category-btn.active .category-count {
		background: rgba(99, 102, 241, 0.2);
		color: #6ee7b7;
	}

	:global(.dark) .your-apps-section {
		border-color: #333;
	}

	:global(.dark) .your-apps-count {
		background: rgba(34, 197, 94, 0.1);
		color: #4ade80;
	}

	:global(.dark) .main-content {
		background: #1a1a1a;
	}

	:global(.dark) .search-wrapper {
		background: #262626;
		border-color: #333;
	}

	:global(.dark) .search-wrapper:focus-within {
		border-color: #10B981;
		background: #1f1f1f;
	}

	:global(.dark) .search-input {
		color: #ffffff;
	}

	:global(.dark) .search-input::placeholder {
		color: #6b7280;
	}

	:global(.dark) .section-header h3 {
		color: #ffffff;
	}

	:global(.dark) .app-count {
		color: #9ca3af;
	}

	:global(.dark) .app-card {
		background: #262626 !important;
		border-color: #333 !important;
	}

	:global(.dark) .app-card:hover {
		background: #2d2d2d !important;
		border-color: #444 !important;
	}

	:global(.dark) .app-card.added {
		background: rgba(34, 197, 94, 0.1) !important;
		border-color: rgba(34, 197, 94, 0.3) !important;
	}

	:global(.dark) .app-name {
		color: #ffffff !important;
	}

	:global(.dark) .app-desc,
	:global(.dark) .app-description {
		color: #9ca3af !important;
	}

	:global(.dark) .app-logo {
		background: #333 !important;
	}

	:global(.dark) .add-btn {
		background: #333 !important;
		color: #d1d5db !important;
	}

	:global(.dark) .add-btn:hover {
		background: #444 !important;
		color: #ffffff !important;
	}

	:global(.dark) .added-badge {
		background: #22c55e !important;
		color: white !important;
	}

	:global(.dark) .remove-badge {
		background: #ef4444 !important;
		color: white !important;
	}

	/* Dark mode for subcategory headers */
	:global(.dark) .subcategory-header {
		border-color: #333 !important;
	}

	:global(.dark) .subcategory-name {
		color: #d1d5db !important;
	}

	:global(.dark) .subcategory-count {
		background: #333 !important;
		color: #9ca3af !important;
	}

	/* Dark mode for apps grid section */
	:global(.dark) .apps-section {
		background: transparent !important;
	}

	:global(.dark) .apps-grid {
		background: transparent !important;
	}

	:global(.dark) .custom-form {
		background: #1f1f1f;
	}

	:global(.dark) .form-title {
		color: #ffffff;
	}

	:global(.dark) .form-description {
		color: #9ca3af;
	}

	:global(.dark) .form-group label {
		color: #d1d5db;
	}

	:global(.dark) .form-input,
	:global(.dark) .form-select,
	:global(.dark) .form-textarea {
		background: #262626;
		border-color: #333;
		color: #ffffff;
	}

	:global(.dark) .form-input:focus,
	:global(.dark) .form-select:focus,
	:global(.dark) .form-textarea:focus {
		border-color: #10B981;
		background: #1f1f1f;
	}

	:global(.dark) .form-input::placeholder,
	:global(.dark) .form-textarea::placeholder {
		color: #6b7280;
	}

	:global(.dark) .empty-state {
		color: #9ca3af;
	}

	:global(.dark) .empty-state p {
		color: #6b7280;
	}

	:global(.dark) .subcategory-header {
		border-color: #333;
	}

	:global(.dark) .subcategory-name {
		color: #e5e7eb;
	}

	:global(.dark) .subcategory-count {
		background: #333;
		color: #9ca3af;
	}

	/* ===== TAB BADGE ===== */
	.tab-badge {
		background: #10B981;
		color: white;
		font-size: 0.6875rem;
		font-weight: 600;
		padding: 0.125rem 0.4rem;
		border-radius: 9999px;
		min-width: 18px;
		text-align: center;
	}

	.main-tab.active .tab-badge {
		background: rgba(255, 255, 255, 0.25);
	}

	/* ===== REMOVE ICON ON HOVER ===== */
	.remove-badge,
	.remove-icon {
		display: none;
	}

	.app-card.added:hover .added-badge {
		display: none;
	}

	.app-card.added:hover .remove-badge {
		display: flex;
		width: 24px;
		height: 24px;
		align-items: center;
		justify-content: center;
		background: #ef4444;
		border-radius: 0.375rem;
		color: white;
	}

	.featured-card .added-icon {
		display: block;
	}

	.featured-card .remove-icon {
		display: none;
	}

	.featured-card.added:hover .added-icon {
		display: none;
	}

	.featured-card.added:hover .remove-icon {
		display: block;
		color: #ef4444;
	}

	/* Make added cards clickable with hover effect */
	.app-card.added {
		cursor: pointer;
	}

	.app-card.added:hover {
		background: #fef2f2;
		border-color: #fecaca;
	}

	.featured-card.added:hover {
		background: #fef2f2;
		border-color: #fecaca;
	}

	/* ===== MY APPS SECTION ===== */
	.myapps-wrapper {
		height: 100%;
		background: white;
		overflow-y: auto;
	}

	.myapps-empty {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		height: 100%;
		text-align: center;
		color: #9ca3af;
		padding: 2rem;
	}

	.myapps-empty h3 {
		font-size: 1.125rem;
		font-weight: 600;
		color: #374151;
		margin: 1rem 0 0.375rem;
	}

	.myapps-empty p {
		font-size: 0.875rem;
		color: #9ca3af;
		margin: 0 0 1.5rem;
	}

	.browse-btn {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.625rem 1.25rem;
		background: linear-gradient(135deg, #10B981 0%, #059669 100%);
		color: white;
		border: none;
		border-radius: 0.5rem;
		font-size: 0.875rem;
		font-weight: 500;
		cursor: pointer;
		transition: all 0.15s;
	}

	.browse-btn:hover {
		transform: translateY(-1px);
		box-shadow: 0 4px 12px -2px rgba(16, 185, 129, 0.4);
	}

	.myapps-header {
		padding: 1.5rem;
		border-bottom: 1px solid #e5e7eb;
	}

	.myapps-header h3 {
		font-size: 1rem;
		font-weight: 600;
		color: #111827;
		margin: 0 0 0.25rem;
	}

	.myapps-header p {
		font-size: 0.8125rem;
		color: #6b7280;
		margin: 0;
	}

	.myapps-list {
		padding: 1rem;
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.myapp-item {
		display: flex;
		align-items: center;
		gap: 0.875rem;
		padding: 0.875rem 1rem;
		background: #fafafa;
		border: 1px solid #e5e7eb;
		border-radius: 0.5rem;
		transition: all 0.15s;
	}

	.myapp-item:hover {
		background: #f5f5f5;
		border-color: #d1d5db;
	}

	.myapp-item.disabled {
		opacity: 0.6;
		background: #f9fafb;
	}

	.myapp-logo {
		width: 40px;
		height: 40px;
		border-radius: 0.5rem;
		overflow: hidden;
		background: white;
		flex-shrink: 0;
		border: 1px solid #e5e7eb;
	}

	.myapp-logo img {
		width: 100%;
		height: 100%;
		object-fit: contain;
	}

	.myapp-logo-placeholder {
		width: 100%;
		height: 100%;
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 1rem;
		font-weight: 600;
		color: white;
	}

	.myapp-info {
		flex: 1;
		min-width: 0;
	}

	.myapp-name {
		display: block;
		font-size: 0.875rem;
		font-weight: 500;
		color: #111827;
	}

	.myapp-desc,
	.myapp-url {
		display: block;
		font-size: 0.75rem;
		color: #6b7280;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.myapp-actions {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		flex-shrink: 0;
	}

	.myapp-toggle {
		display: flex;
		align-items: center;
		justify-content: center;
		background: none;
		border: none;
		padding: 0.25rem;
		cursor: pointer;
		color: #d1d5db;
		transition: all 0.15s;
	}

	.myapp-toggle.active {
		color: #10B981;
	}

	.myapp-toggle:hover {
		transform: scale(1.1);
	}

	.myapp-remove {
		display: flex;
		align-items: center;
		justify-content: center;
		background: #f3f4f6;
		border: none;
		border-radius: 0.375rem;
		padding: 0.5rem;
		cursor: pointer;
		color: #9ca3af;
		transition: all 0.15s;
	}

	.myapp-remove:hover {
		background: #fee2e2;
		color: #ef4444;
	}

	/* ===== MY APPS DARK MODE ===== */
	:global(.dark) .tab-badge {
		background: rgba(16, 185, 129, 0.8);
	}

	:global(.dark) .main-tab.active .tab-badge {
		background: rgba(255, 255, 255, 0.25);
	}

	:global(.dark) .app-card.added:hover {
		background: rgba(239, 68, 68, 0.1);
		border-color: rgba(239, 68, 68, 0.3);
	}

	:global(.dark) .featured-card.added:hover {
		background: rgba(239, 68, 68, 0.1);
		border-color: rgba(239, 68, 68, 0.3);
	}

	:global(.dark) .myapps-wrapper {
		background: #1a1a1a;
	}

	:global(.dark) .myapps-empty h3 {
		color: #ffffff;
	}

	:global(.dark) .myapps-empty p {
		color: #9ca3af;
	}

	:global(.dark) .myapps-header {
		border-color: #333;
	}

	:global(.dark) .myapps-header h3 {
		color: #ffffff;
	}

	:global(.dark) .myapps-header p {
		color: #9ca3af;
	}

	:global(.dark) .myapp-item {
		background: #262626;
		border-color: #333;
	}

	:global(.dark) .myapp-item:hover {
		background: #2d2d2d;
		border-color: #444;
	}

	:global(.dark) .myapp-item.disabled {
		background: #1f1f1f;
		opacity: 0.5;
	}

	:global(.dark) .myapp-logo {
		background: #333;
		border-color: #444;
	}

	:global(.dark) .myapp-name {
		color: #ffffff;
	}

	:global(.dark) .myapp-desc,
	:global(.dark) .myapp-url {
		color: #9ca3af;
	}

	:global(.dark) .myapp-toggle {
		color: #6b7280;
	}

	:global(.dark) .myapp-toggle.active {
		color: #34d399;
	}

	:global(.dark) .myapp-remove {
		background: #333;
		color: #9ca3af;
	}

	:global(.dark) .myapp-remove:hover {
		background: rgba(239, 68, 68, 0.2);
		color: #f87171;
	}

	/* ===== PAGE MODE DARK STYLES ===== */
	:global(.dark) .page-container {
		background: #1a1a1a;
	}

	:global(.dark) .page-content {
		background: #1a1a1a;
	}

	:global(.dark) .page-header {
		background: #1f1f1f;
		border-color: #333;
	}

	:global(.dark) .page-header .header-text h2 {
		color: #ffffff;
	}

	:global(.dark) .page-header .header-text p {
		color: #9ca3af;
	}

	:global(.dark) .page-container .main-tabs {
		background: #1f1f1f;
		border-color: #333;
	}

	:global(.dark) .page-container .main-tab {
		color: #9ca3af;
		background: #333;
		border-color: #444;
	}

	:global(.dark) .page-container .main-tab:hover {
		background: #444;
		color: #ffffff;
		border-color: #555;
	}

	:global(.dark) .page-container .main-tab.active {
		background: #0A84FF !important;
		color: white !important;
		border-color: #0A84FF !important;
	}

	:global(.dark) .page-container .sidebar {
		background: #1f1f1f;
		border-color: #333;
	}

	:global(.dark) .page-container .sidebar-title {
		color: #9ca3af;
	}

	:global(.dark) .page-container .category-btn {
		color: #d1d5db;
		background: transparent;
	}

	:global(.dark) .page-container .category-btn:hover {
		background: #333;
		color: #ffffff;
	}

	:global(.dark) .page-container .category-btn.active {
		background: rgba(10, 132, 255, 0.2) !important;
		color: #0A84FF !important;
	}

	:global(.dark) .page-container .category-count {
		color: #6b7280;
	}

	:global(.dark) .page-container .main-content {
		background: #1a1a1a;
	}

	:global(.dark) .page-container .search-wrapper {
		background: #262626;
		border-color: #333;
	}

	:global(.dark) .page-container .search-input {
		color: #ffffff;
	}

	:global(.dark) .page-container .search-input::placeholder {
		color: #6b7280;
	}

	:global(.dark) .page-container .featured-section {
		background: #262626 !important;
		border: 1px solid #333 !important;
	}

	:global(.dark) .page-container .section-header h3 {
		color: #ffffff !important;
	}

	:global(.dark) .page-container .section-icon {
		color: #0A84FF !important;
	}

	:global(.dark) .page-container .featured-card {
		background: #333 !important;
		border: 1px solid #444 !important;
	}

	:global(.dark) .page-container .featured-card:hover:not(:disabled) {
		background: #3d3d3d !important;
		border-color: #555 !important;
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3) !important;
	}

	:global(.dark) .page-container .featured-card.added {
		background: rgba(34, 197, 94, 0.15) !important;
		border-color: rgba(34, 197, 94, 0.3) !important;
	}

	:global(.dark) .page-container .featured-name {
		color: #ffffff !important;
	}

	:global(.dark) .page-container .featured-desc {
		color: #9ca3af !important;
	}

	:global(.dark) .page-container .featured-logo {
		background: #262626 !important;
	}

	:global(.dark) .page-container .featured-action button {
		background: #444 !important;
		color: #d1d5db !important;
		border-color: #555 !important;
	}

	:global(.dark) .page-container .featured-action button:hover {
		background: #555 !important;
		color: #ffffff !important;
	}

	:global(.dark) .page-container .app-card {
		background: #262626 !important;
		border-color: #333 !important;
	}

	:global(.dark) .page-container .app-card:hover {
		background: #2d2d2d !important;
		border-color: #444 !important;
	}

	:global(.dark) .page-container .subcategory-header {
		border-color: #333 !important;
	}

	:global(.dark) .page-container .subcategory-name {
		color: #d1d5db !important;
	}

	:global(.dark) .page-container .subcategory-count {
		background: #333 !important;
		color: #9ca3af !important;
	}

	:global(.dark) .page-container .apps-section .section-header {
		border-bottom-color: #333 !important;
	}

	:global(.dark) .page-container .apps-section .section-header h3 {
		color: #ffffff;
	}

	:global(.dark) .page-container .section-count,
	:global(.dark) .page-container .app-count {
		color: #9ca3af !important;
	}

	:global(.dark) .page-container .your-apps-count {
		background: #333 !important;
		color: #d1d5db !important;
		border-color: #444 !important;
	}

	:global(.dark) .page-container .your-apps-section {
		border-top-color: #333 !important;
	}

	:global(.dark) .page-container .your-apps-section .sidebar-title {
		color: #9ca3af;
	}

	:global(.dark) .page-container .myapps-wrapper {
		background: #1a1a1a;
	}

	:global(.dark) .page-container .myapps-header h3 {
		color: #ffffff;
	}

	:global(.dark) .page-container .myapps-header p {
		color: #9ca3af;
	}

	:global(.dark) .page-container .myapp-card {
		background: #262626;
		border-color: #333;
	}

	:global(.dark) .page-container .myapp-card:hover {
		background: #2d2d2d;
		border-color: #444;
	}

	:global(.dark) .page-container .myapp-name {
		color: #ffffff;
	}

	:global(.dark) .page-container .myapp-url {
		color: #9ca3af;
	}

	:global(.dark) .page-container .myapp-category {
		color: #9ca3af;
	}

	:global(.dark) .page-container .myapp-actions button {
		background: #333;
		color: #d1d5db;
	}

	:global(.dark) .page-container .myapp-actions button:hover {
		background: #444;
		color: #ffffff;
	}

	:global(.dark) .page-container .app-name {
		color: #ffffff;
	}

	:global(.dark) .page-container .app-description {
		color: #9ca3af;
	}

	:global(.dark) .page-container .add-btn {
		background: #333;
		color: #d1d5db;
	}

	:global(.dark) .page-container .add-btn:hover {
		background: #444;
		color: #ffffff;
	}

	/* Missing dark mode styles for content areas */
	:global(.dark) .page-container .content-scroll {
		background: #1a1a1a !important;
	}

	:global(.dark) .page-container .browse-layout {
		background: #1a1a1a !important;
	}

	:global(.dark) .page-container .apps-grid {
		background: transparent !important;
	}

	:global(.dark) .page-container .featured-grid {
		background: transparent !important;
	}

	/* Ensure base containers override light mode */
	:global(.dark) .page-container {
		background: #1a1a1a !important;
	}

	:global(.dark) .page-content {
		background: #1a1a1a !important;
	}

	:global(.dark) .page-container .main-content {
		background: #1a1a1a !important;
	}

	/* My Apps wrapper in dark mode */
	:global(.dark) .page-container .myapps-wrapper {
		background: #1a1a1a !important;
	}

	:global(.dark) .page-container .myapps-empty {
		background: transparent !important;
		color: #9ca3af !important;
	}

	:global(.dark) .page-container .myapps-list {
		background: transparent !important;
	}

	:global(.dark) .page-container .myapp-item {
		background: #262626 !important;
		border-color: #333 !important;
	}

	:global(.dark) .page-container .myapp-item:hover {
		background: #2d2d2d !important;
		border-color: #444 !important;
	}
</style>
