// mixed_test.js
// 混合场景压测脚本
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter, Trend } from 'k6/metrics';

const BASE_URL = 'http://localhost:2333';

const readOps = new Counter('read_operations');
const writeOps = new Counter('write_operations');
const authOps = new Counter('auth_operations');
const responseTime = new Trend('custom_response_time');

export function setup() {
  // 预先注册一些测试用户
  for (let i = 0; i < 10; i++) {
    const username = `mixedtest_user_${i}`;
    http.post(`${BASE_URL}/api/auth/register`, 
      JSON.stringify({ 
        username: username, 
        password: '123456',
        env: { ipv4: '192.168.1.1', os: 'Windows', device_info: 'Chrome' }
      }),
      { headers: { 'Content-Type': 'application/json' } }
    );
  }

  // 登录获取Token和cookie
  const loginRes = http.post(`${BASE_URL}/api/auth/login`,
    JSON.stringify({ 
      username: 'mixedtest_user_0', 
      password: '123456',
      env: { ipv4: '192.168.1.1', os: 'Windows', device_info: 'Chrome' }
    }),
    { headers: { 'Content-Type': 'application/json' } }
  );
  
  // 从cookie中获取token
  let accessToken = '';
  let refreshToken = '';
  const cookies = loginRes.cookies;
  if (cookies['access-token'] && cookies['access-token'].length > 0) {
    accessToken = cookies['access-token'][0].value;
  }
  if (cookies['refresh-token'] && cookies['refresh-token'].length > 0) {
    refreshToken = cookies['refresh-token'][0].value;
  }
  
  // 从响应中获取token（备用）
  try {
    const body = loginRes.json();
    if (body.data && body.data.token) {
      if (!accessToken) accessToken = body.data.token.access_token || '';
      if (!refreshToken) refreshToken = body.data.token.refresh_token || '';
    }
  } catch (e) {
    console.log('Failed to parse login response:', e);
  }

  console.log('Setup complete');
  console.log('Access token:', accessToken.substring(0, 50) + '...');
  
  return { 
    accessToken: accessToken,
    refreshToken: refreshToken
  };
}

export const options = {
  stages: [
    { duration: '60s', target: 50 },    // 逐步加压到50用户
    { duration: '120s', target: 50 },   // 保持50用户120秒
    { duration: '60s', target: 100 },   // 加压到100用户
    { duration: '120s', target: 100 },  // 保持100用户120秒
    { duration: '60s', target: 0 },     // 逐步降压
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'],   // 95%请求响应时间<500ms
    http_req_failed: ['rate<0.02'],     // 错误率<2%
  },
};

export default function (data) {
  // 手动构造Cookie头
  const cookieHeader = `access-token=${data.accessToken}; refresh-token=${data.refreshToken}`;
  
  const params = {
    headers: { 
      'Content-Type': 'application/json',
      'Cookie': cookieHeader,
    },
  };

  const rand = Math.random();

  if (rand < 0.5) {
    // 50% 读操作
    readOps.add(1);
    
    const readRand = Math.random();
    if (readRand < 0.4) {
      // 读帖子列表
      const page = Math.floor(Math.random() * 10) + 1;
      const res = http.get(`${BASE_URL}/api/post?page=${page}&page_size=10`, params);
      check(res, { 'list status 200': (r) => r.status === 200 });
      responseTime.add(res.timings.duration);
    } else if (readRand < 0.7) {
      // 搜索帖子
      const keywords = ['test', 'demo', 'hello', 'golang', 'forum', 'trading'];
      const keyword = keywords[Math.floor(Math.random() * keywords.length)];
      const res = http.get(`${BASE_URL}/api/post/search?keyword=${keyword}`, params);
      check(res, { 'search status 200': (r) => r.status === 200 });
      responseTime.add(res.timings.duration);
    } else {
      // 获取行情（无需认证）
      const res = http.get(`${BASE_URL}/api/market`);
      check(res, { 'market status 200': (r) => r.status === 200 });
      responseTime.add(res.timings.duration);
    }
  } else if (rand < 0.8) {
    // 30% 写操作
    writeOps.add(1);
    
    const payload = JSON.stringify({
      title: `Load Test Post ${Date.now()}`,
      content: 'This is a load test post content',
      category: 'test',
      tags: ['loadtest', 'performance']
    });
    
    const res = http.post(`${BASE_URL}/api/create`, payload, params);
    check(res, { 'create status 200': (r) => r.status === 200 || r.status === 201 });
    responseTime.add(res.timings.duration);
  } else {
    // 20% 认证操作
    authOps.add(1);
    
    const userIndex = Math.floor(Math.random() * 10);
    const username = `mixedtest_user_${userIndex}`;
    
    const loginPayload = JSON.stringify({ 
      username: username, 
      password: '123456',
      env: { ipv4: '192.168.1.1', os: 'Windows', device_info: 'Chrome' }
    });
    
    const res = http.post(`${BASE_URL}/api/auth/login`, loginPayload, {
      headers: { 'Content-Type': 'application/json' }
    });
    check(res, { 'login status 200': (r) => r.status === 200 });
    responseTime.add(res.timings.duration);
  }

  sleep(0.5);
}
