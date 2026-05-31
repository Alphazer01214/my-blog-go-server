// register_test.js
// 注册接口压测脚本
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter } from 'k6/metrics';

const successfulRegistrations = new Counter('successful_registrations');

export const options = {
  stages: [
    { duration: '30s', target: 20 },  // 逐步加压到20用户
    { duration: '60s', target: 20 },  // 保持20用户60秒
    { duration: '30s', target: 0 },   // 逐步降压
  ],
  thresholds: {
    http_req_duration: ['p(95)<1000'], // 95%请求响应时间<1000ms
    http_req_failed: ['rate<0.05'],    // 错误率<5%
  },
};

let userCounter = 0;

export default function () {
  userCounter++;
  const username = `loadtest_user_${__VU}_${userCounter}`;
  
  const payload = JSON.stringify({
    username: username,
    password: '123456',
    env: {
      ipv4: '192.168.1.1',
      ipv6: '',
      os: 'Windows 11',
      device_info: 'Chrome 120'
    }
  });

  const params = {
    headers: { 'Content-Type': 'application/json' },
  };

  const res = http.post('http://localhost:2333/api/auth/register', payload, params);
  
  const success = check(res, {
    'status is 200': (r) => r.status === 200,
    'response time < 1000ms': (r) => r.timings.duration < 1000,
  });

  if (success) {
    successfulRegistrations.add(1);
  }

  sleep(1);
}
