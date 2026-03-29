import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// Custom metrics
const enduranceErrorRate = new Rate('endurance_errors');
const enduranceResponseTime = new Trend('endurance_response_time');
const memoryLeakIndicator = new Trend('memory_leak_indicator');

// Test configuration
export const options = {
  stages: [
    { duration: '2h', target: 200 },  // Sustained load: 200 req/s for 2 hours
    { duration: '5m', target: 0 },     // Cool down
  ],
  thresholds: {
    // RED PHASE: These thresholds WILL FAIL due to memory leaks or degradation
    'endurance_errors': ['rate<0.01'],           // Error rate must be < 1%
    'http_req_duration': ['p(95)<500'],          // P95 must be < 500ms
    'http_req_duration': ['p(99)<1000'],         // P99 must be < 1000ms
    'endurance_response_time': ['p(95)<500'],    // Endurance P95 < 500ms
    'memory_leak_indicator': ['avg<100'],        // Memory growth indicator
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8001';
let requestCount = 0;

export function setup() {
  // Verify BusinessOS endpoint is accessible
  const response = http.get(`${BASE_URL}/healthz`);
  if (response.status !== 200) {
    throw new Error(`BusinessOS health check failed: ${response.status}`);
  }
  return { baseURL: BASE_URL, startTime: Date.now() };
}

export default function(data) {
  requestCount++;

  // Test public BusinessOS endpoints over sustained period
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
  enduranceResponseTime.add(response.timings.duration);

  // Memory leak detection: track response time degradation
  const elapsedMinutes = (Date.now() - data.startTime) / 60000;
  memoryLeakIndicator.add(response.timings.duration * elapsedMinutes);

  // Check success conditions
  const success = check(response, {
    'status is 2xx': (r) => r.status >= 200 && r.status < 300,
    'response time < 1000ms': (r) => r.timings.duration < 1000,
    'response body is valid': (r) => r.body.length > 0,
    'no memory degradation': (r) => {
      // Fail if response time increases by >50% over time
      const avgResponseTime = enduranceResponseTime.avg;
      return r.timings.duration < avgResponseTime * 1.5;
    },
  });

  enduranceErrorRate.add(!success);

  // Log progress every 1000 requests
  if (requestCount % 1000 === 0) {
    console.log(`Progress: ${requestCount} requests, elapsed: ${elapsedMinutes.toFixed(2)}min`);
    console.log(`Avg response time: ${enduranceResponseTime.avg.toFixed(2)}ms`);
  }

  // Consistent pacing
  sleep(2);
}

export function teardown(data) {
  const elapsedMinutes = (Date.now() - data.startTime) / 60000;
  console.log('Endurance test completed');
  console.log(`Total requests: ${requestCount}`);
  console.log(`Duration: ${elapsedMinutes.toFixed(2)} minutes`);
  console.log(`Final avg response time: ${enduranceResponseTime.avg.toFixed(2)}ms`);
}
