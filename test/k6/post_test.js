// post_test.js
// 帖子接口压测脚本（需要先登录获取Token）
import http from 'k6/http';
import { check, sleep } from 'k6';

const BASE_URL = 'http://localhost:2333';

// 先登录获取Token（cookie会自动保存到cookie jar）
export function setup() {
  const loginPayload = JSON.stringify({
    username: 'testuser',
    password: '123456',
    env: {
      ipv4: '192.168.1.1',
      ipv6: '',
      os: 'Windows 11',
      device_info: 'Chrome 120'
    }
  });

  const loginRes = http.post(`${BASE_URL}/api/auth/login`, loginPayload, {
    headers: { 'Content-Type': 'application/json' },
  });
  
  if (loginRes.status !== 200) {
    throw new Error('Login failed: ' + loginRes.body);
  }

  // 从响应中获取token
  let accessToken = '';
  let refreshToken = '';
  try {
    const body = loginRes.json();
    if (body.data && body.data.token) {
      accessToken = body.data.token.access_token || '';
      refreshToken = body.data.token.refresh_token || '';
    }
  } catch (e) {
    console.log('Failed to parse login response:', e);
  }

  // 从cookie中获取token（更可靠）
  const cookies = loginRes.cookies;
  if (cookies['access-token'] && cookies['access-token'].length > 0) {
    accessToken = cookies['access-token'][0].value;
  }
  if (cookies['refresh-token'] && cookies['refresh-token'].length > 0) {
    refreshToken = cookies['refresh-token'][0].value;
  }
  
  console.log('Login successful');
  console.log('Access token:', accessToken.substring(0, 50) + '...');
  console.log('Refresh token:', refreshToken.substring(0, 50) + '...');
  
  return { 
    accessToken: accessToken,
    refreshToken: refreshToken
  };
}

export const options = {
  stages: [
    { duration: '30s', target: 50 },   // 逐步加压到50用户
    { duration: '60s', target: 50 },   // 保持50用户60秒
    { duration: '30s', target: 100 },  // 加压到100用户
    { duration: '60s', target: 100 },  // 保持100用户60秒
    { duration: '30s', target: 0 },    // 逐步降压
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'],  // 95%请求响应时间<500ms
    http_req_failed: ['rate<0.01'],    // 错误率<1%
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

  // 随机选择操作
  const rand = Math.random();

  if (rand < 0.6) {
    // 60% 读帖子列表
    const page = Math.floor(Math.random() * 10) + 1;
    const res = http.get(`${BASE_URL}/api/post?page=${page}&page_size=10`, params);
    
    // 调试：打印第一个请求的响应
    if (__ITER === 0 && __VU === 1) {
      console.log('Cookie header:', cookieHeader);
      console.log('List response status:', res.status);
      console.log('List response body:', res.body);
    }
    
    check(res, { 
      'list status 200': (r) => r.status === 200,
      'list has data': (r) => {
        try {
          const body = r.json();
          return body.code === 0;
        } catch (e) {
          return false;
        }
      }
    });
  } else if (rand < 0.8) {
    // 20% 搜索帖子
    const keywords = ['test', 'demo', 'hello', 'golang', 'forum'];
    const keyword = keywords[Math.floor(Math.random() * keywords.length)];
    const res = http.get(`${BASE_URL}/api/post/search?keyword=${keyword}&page=1&page_size=10`, params);
    
    check(res, { 
      'search status 200': (r) => r.status === 200,
      'search has data': (r) => {
        try {
          const body = r.json();
          return body.code === 0;
        } catch (e) {
          return false;
        }
      }
    });
  } else {
    // 20% 获取帖子详情
    const postId = Math.floor(Math.random() * 100) + 1;
    const res = http.get(`${BASE_URL}/api/post/${postId}`, params);
    
    check(res, { 
      'detail status 200': (r) => r.status === 200 || r.status === 404,
    });
  }

  sleep(0.5);
}
