// sse_test.js
// SSE流式推送压测脚本
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '30s', target: 50 },   // 逐步加压到50连接
    { duration: '60s', target: 50 },   // 保持50连接60秒
    { duration: '30s', target: 100 },  // 加压到100连接
    { duration: '60s', target: 100 },  // 保持100连接60秒
    { duration: '30s', target: 0 },    // 逐步降压
  ],
  thresholds: {
    http_req_duration: ['p(95)<1000'],  // 95%请求响应时间<1000ms
    http_req_failed: ['rate<0.05'],     // 错误率<5%
  },
};

export default function () {
  const url = 'http://localhost:2333/api/market/stream';
  const params = {
    headers: { 
      'Accept': 'text/event-stream',
      'Cache-Control': 'no-cache'
    },
    timeout: '60s',
  };

  const res = http.get(url, params);
  
  check(res, {
    'status is 200': (r) => r.status === 200,
    'is SSE stream': (r) => {
      const contentType = r.headers['Content-Type'] || r.headers['content-type'];
      return contentType && contentType.includes('text/event-stream');
    },
    'has data': (r) => r.body && r.body.length > 0,
  });

  // 保持连接一段时间
  sleep(5);
}
