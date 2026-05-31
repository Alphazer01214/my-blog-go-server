// market_test.js
// 行情数据接口压测脚本
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '30s', target: 100 },  // 逐步加压到100用户
    { duration: '60s', target: 100 },  // 保持100用户60秒
    { duration: '30s', target: 200 },  // 加压到200用户
    { duration: '60s', target: 200 },  // 保持200用户60秒
    { duration: '30s', target: 0 },    // 逐步降压
  ],
  thresholds: {
    http_req_duration: ['p(95)<300'],  // 95%请求响应时间<300ms
    http_req_failed: ['rate<0.01'],    // 错误率<1%
  },
};

export default function () {
  const res = http.get('http://localhost:2333/api/market');
  
  check(res, {
    'status is 200': (r) => r.status === 200,
    'has market data': (r) => {
      try {
        const json = r.json();
        return json.code === 0;
      } catch (e) {
        return false;
      }
    },
  });

  sleep(0.5);
}
