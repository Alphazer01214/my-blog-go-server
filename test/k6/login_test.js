// login_test.js
// 登录接口压测脚本
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '30s', target: 50 },  // 逐步加压到50用户
    { duration: '60s', target: 50 },  // 保持50用户60秒
    { duration: '30s', target: 100 }, // 加压到100用户
    { duration: '60s', target: 100 }, // 保持100用户60秒
    { duration: '30s', target: 0 },   // 逐步降压
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95%请求响应时间<500ms
    http_req_failed: ['rate<0.01'],   // 错误率<1%
  },
};

export default function () {
  const payload = JSON.stringify({
    username: 'testuser',
    password: '123456',
  });

  const params = {
    headers: { 'Content-Type': 'application/json' },
  };

  const res = http.post('http://localhost:2333/api/auth/login', payload, params);
  
  check(res, {
    'status is 200': (r) => r.status === 200,
    'response time < 500ms': (r) => r.timings.duration < 500,
  });

  sleep(1);
}
