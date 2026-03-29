import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const osaResponseTime = new Trend('osa_response_time');

// Test configuration
export const options = {
  stages: [
    { duration: '5m', target: 100 },   // Ramp up to 100 req/s
    { duration: '5m', target: 500 },   // Ramp up to 500 req/s
    { duration: '5m', target: 1000 },  // Ramp up to 1000 req/s
    { duration: '2m', target: 0 },     // Cool down
  ],
  thresholds: {
    // RED PHASE: These thresholds WILL FAIL initially
    'errors': ['rate<0.01'],            // Error rate must be < 1%
    'http_req_duration': ['p(95)<500'], // P95 latency must be < 500ms
    'http_req_duration': ['p(99)<1000'], // P99 latency must be < 1000ms
    'osa_response_time': ['p(95)<500'], // OSA endpoint P95 < 500ms
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8001';

export function setup() {
  // Warmup: verify BusinessOS endpoint is accessible
  const response = http.get(`${BASE_URL}/healthz`);
  if (response.status !== 200) {
    throw new Error(`BusinessOS health check failed: ${response.status}`);
  }
  return { baseURL: BASE_URL };
}

export default function(data) {
  // Test public BusinessOS endpoints (no auth required)
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
  osaResponseTime.add(response.timings.duration);

  // Check success conditions
  const success = check(response, {
    'status is 2xx': (r) => r.status >= 200 && r.status < 300,
    'response time < 500ms': (r) => r.timings.duration < 500,
    'response body is valid': (r) => r.body.length > 0,
  });

  errorRate.add(!success);

  // Small pause between requests
  sleep(Math.random() * 2 + 1); // 1-3 seconds
}

export function teardown(data) {
  console.log('Load test completed');
  console.log(`Total errors: ${errorRate.name}`);
}
