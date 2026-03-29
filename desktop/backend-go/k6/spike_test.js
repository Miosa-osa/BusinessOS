import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

// Custom metrics
const spikeErrorRate = new Rate('spike_errors');
const spikeResponseTime = new Trend('spike_response_time');

// Test configuration
export const options = {
  stages: [
    { duration: '1m', target: 100 },    // Normal load: 100 req/s
    { duration: '30s', target: 5000 },  // SPIKE: 5000 req/s (50x increase)
    { duration: '1m', target: 100 },    // Recovery: back to 100 req/s
    { duration: '1m', target: 0 },      // Cool down
  ],
  thresholds: {
    // RED PHASE: These thresholds WILL FAIL during spike
    'spike_errors': ['rate<0.05'],           // Error rate must be < 5% during spike
    'http_req_duration': ['p(95)<1000'],     // P95 must be < 1000ms
    'http_req_duration': ['p(99)<2000'],     // P99 must be < 2000ms
    'spike_response_time': ['p(95)<1000'],   // Spike P95 < 1000ms
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
  // Test public BusinessOS endpoints under spike conditions
  const endpoints = [
    { path: '/healthz', method: 'GET', name: 'Healthz' },
    { path: '/api/yawl/health', method: 'GET', name: 'YAWLHealth' },
    { path: '/api/pm4py/health', method: 'GET', name: 'Pm4pyHealth' },
  ];

  const endpoint = endpoints[Math.floor(Math.random() * endpoints.length)];
  let url = `${data.baseURL}${endpoint.path}`;
  let params = {
    tags: { name: endpoint.name },
    headers: {
      'Content-Type': 'application/json',
    },
  };

  let response = http.get(url, params);

  // Track custom metrics
  spikeResponseTime.add(response.timings.duration);

  // Check success conditions
  const success = check(response, {
    'status is 2xx': (r) => r.status >= 200 && r.status < 300,
    'response time < 2000ms': (r) => r.timings.duration < 2000,
    'no timeout errors': (r) => r.error === null || r.error_code !== 13, // 13 = socket timeout
  });

  spikeErrorRate.add(!success);

  // Minimal pause during spike
  sleep(0.1); // 100ms
}

export function teardown(data) {
  console.log('Spike test completed');
  console.log(`Spike errors: ${spikeErrorRate.name}`);
}
