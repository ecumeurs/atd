/**
 * Health Module — Monitors LLM provider and task availability
 * @spec-link [[ui_webui_health_indicator]]
 */
import { fetchHealth } from './api.js';

const healthIcon = document.getElementById('health-icon');
const healthStatus = document.getElementById('health-status');
const healthIndicator = document.getElementById('health-indicator');
const healthModal = document.getElementById('health-modal');
const healthContent = document.getElementById('health-content');
const closeHealthModal = document.getElementById('close-health-modal');

let lastHealth = null;

export function initHealth() {
    // Click to show details
    healthIndicator?.addEventListener('click', showHealthModal);
    closeHealthModal?.addEventListener('click', () => {
        healthModal.style.display = 'none';
    });

    // Initial check
    checkHealth();

    // Periodic refresh (every 30 seconds)
    setInterval(checkHealth, 30000);
}

async function checkHealth() {
    try {
        const health = await fetchHealth();
        lastHealth = health;
        updateIndicator(health);
    } catch (err) {
        console.error('Health check failed:', err);
        updateIndicatorError(err.message);
    }
}

function updateIndicator(health) {
    if (!health || !health.providers) {
        setIndicatorStatus('unknown', 'Unknown');
        return;
    }

    // Count available providers (exclude passthrough)
    const providers = health.providers.filter(p => p.type !== 'passthrough');
    const available = providers.filter(p => p.status === 'available').length;

    if (available === providers.length) {
        setIndicatorStatus('good', 'All Ready');
    } else if (available > 0) {
        setIndicatorStatus('warning', 'Partial');
    } else {
        setIndicatorStatus('error', 'Offline');
    }
}

function updateIndicatorError(message) {
    setIndicatorStatus('error', 'Error');
}

function setIndicatorStatus(status, text) {
    healthIcon.textContent = {
        'good': '🟢',
        'warning': '🟡',
        'error': '🔴',
        'unknown': '⚪'
    }[status] || '⚪';

    healthStatus.textContent = text;

    healthIndicator.className = 'info-badge health-badge health-' + status;
}

async function showHealthModal() {
    if (!lastHealth) {
        await checkHealth();
    }

    if (!lastHealth) {
        healthContent.innerHTML = '<div style="padding: 20px; text-align: center;">Unable to fetch health information</div>';
        healthModal.style.display = 'flex';
        return;
    }

    let html = '<div class="health-section">';

    // Providers section
    html += '<h3>Providers</h3>';
    html += '<div class="health-providers">';
    for (const provider of lastHealth.providers) {
        const statusIcon = provider.status === 'available' ? '🟢' : '🔴';
        const statusText = provider.status === 'available' ? 'Available' : 'Offline';
        html += `
            <div class="health-provider-item">
                <div class="health-provider-header">
                    <span class="health-provider-name">${provider.name}</span>
                    <span class="health-provider-status">${statusIcon} ${statusText}</span>
                </div>
                <div class="health-provider-details">
                    <div>Type: ${provider.type || 'ollama'}</div>
                    ${provider.base_url ? `<div>URL: ${provider.base_url}</div>` : ''}
                    ${provider.models && provider.models.length > 0 ? `
                        <div>Models: ${provider.models.map(m => m.name).join(', ')}</div>
                    ` : ''}
                </div>
            </div>
        `;
    }
    html += '</div>';

    // Tasks section
    html += '<h3>Task Availability</h3>';
    html += '<div class="health-tasks">';
    for (const [task, provider] of Object.entries(lastHealth.tasks)) {
        const isAvailable = provider && !provider.includes('error') && !provider.includes('ide_fallback');
        const statusIcon = isAvailable ? '🟢' : (provider.includes('ide_fallback') ? '🟡' : '🔴');
        html += `
            <div class="health-task-item">
                <span class="health-task-name">${task}</span>
                <span class="health-task-provider">${statusIcon} ${provider || 'Not configured'}</span>
            </div>
        `;
    }
    html += '</div>';

    html += '</div>';

    healthContent.innerHTML = html;
    healthModal.style.display = 'flex';
}
