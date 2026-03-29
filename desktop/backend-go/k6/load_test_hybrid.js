import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// Custom metrics per endpoint
const osaErrorRate = new Rate('osa_errors');
const businessOsErrorRate = new Rate('business_os_errors');
const hybridResponseTime = new Trend('hybrid_response_time');

const osaResponseTime = new Trend('osa_hybrid_response');
const businessOsResponseTime = new Trend('business_os_hybrid_response');

// Test configuration
export const options = {
  stages: [
    { duration: '5m', target: 200 },   // Mixed load: 200 req/s
    { duration: '5m', target: 400 },   // Ramp up: 400 req/s
    { duration: '5m', target: 600 },   // High load: 600 req/s
    { duration: '2m', target: 0 },     // Cool down
  ],
  thresholds: {
    // RED PHASE: These thresholds WILL FAIL across hybrid endpoints
    'osa_errors': ['rate<0.01'],                 // OSA error rate < 1%
    'business_os_errors': ['rate<0.01'],         // BusinessOS error rate < 1%
    'http_req_duration': ['p(95)<600'],          // Overall P95 < 600ms
    'osa_hybrid_response': ['p(95)<500'],        // OSA P95 < 500ms
    'business_os_hybrid_response': ['p(95)<600'], // BusinessOS P95 < 600ms
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8001';

export function setup() {
  // Verify BusinessOS endpoint is accessible
  const response = http.get(`${BASE_URL}/healthz`);
  if (response.status !== 200) {
    throw new Error(`BusinessOS health check failed: ${response.status}`);
  }
  return { baseURL: BASE_URL };
}

export default function(data) {
  // Mix of public BusinessOS endpoints (all require no auth)
  const endpoints = [
    { path: '/healthz', weight: 0.4, name: 'Healthz' },
    { path: '/api/yawl/health', weight: 0.3, name: 'YAWLHealth' },
    { path: '/api/pm4py/health', weight: 0.3, name: 'Pm4pyHealth' },
  ];

  // Weighted random selection
  const rand = Math.random();
  let cumulative = 0;
  let selectedEndpoint = endpoints[0];

  for (const ep of endpoints) {
    cumulative += ep.weight;
    if (rand <= cumulative) {
      selectedEndpoint = ep;
      break;
    }
  }

  let url = `${data.baseURL}${selectedEndpoint.path}`;
  let params = {
    tags: { name: selectedEndpoint.name },
    headers: { 'Content-Type': 'application/json' },
  };

  let response = http.get(url, params);

  // Track custom metrics
  hybridResponseTime.add(response.timings.duration);

  // Check success conditions
  const success = check(response, {
    'status is 2xx': (r) => r.status >= 200 && r.status < 300,
    'response time < 600ms': (r) => r.timings.duration < 600,
    'response body is valid': (r) => r.body.length > 0 || r.status === 204,
  });

  // Track errors per endpoint type
  if (selectedEndpoint.name === 'Healthz') {
    osaErrorRate.add(!success);
  } else {
    businessOsErrorRate.add(!success);
  }

  // Variable pause to simulate real traffic
  sleep(Math.random() * 3 + 1); // 1-4 seconds
}

export function teardown(data) {
  console.log('Hybrid load test completed');
  console.log(`OSA errors: ${osaErrorRate.name}`);
  console.log(`BusinessOS errors: ${businessOsErrorRate.name}`);
  const avgTime = hybridResponseTime.avg || 0;
  console.log(`Overall avg response time: ${avgTime.toFixed(2)}ms`);
}
