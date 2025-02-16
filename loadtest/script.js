import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '30s', target: 20 }, // Разогрев: повышаем до 20 пользователей за 30 секунд
    { duration: '1m', target: 20 },  // Нагрузка: держим 20 пользователей в течение 1 минуты
    { duration: '30s', target: 50 }, // Повышаем до 50 пользователей за 30 секунд
    { duration: '1m', target: 50 },  // Держим 50 пользователей 1 минуту
    { duration: '30s', target: 0 },  // Плавное снижение нагрузки
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95% запросов должны выполняться быстрее 500мс
    http_req_failed: ['rate<0.1'],    // Менее 10% запросов могут завершиться с ошибкой
  },
};

const BASE_URL = 'http://localhost:8080';
let authToken = '';

export function setup() {
  // Аутентификация для получения токена
  const loginRes = http.post(`${BASE_URL}/api/auth`, JSON.stringify({
    username: 'testuser',
    password: 'testpass'
  }), {
    headers: { 'Content-Type': 'application/json' },
  });
  
  check(loginRes, {
    'успешная аутентификация': (r) => r.status === 200,
  });
  
  const token = loginRes.json('token');
  return { token };
}

export default function(data) {
  const headers = {
    'Authorization': `Bearer ${data.token}`,
    'Content-Type': 'application/json',
  };

  // Тест получения информации
  const infoRes = http.get(`${BASE_URL}/api/info`, { headers });
  check(infoRes, {
    'получение информации успешно': (r) => r.status === 200,
  });

  // Тест отправки монет
  const sendCoinRes = http.post(`${BASE_URL}/api/sendCoin`, 
    JSON.stringify({
      toUser: 'anotheruser',
      amount: 1
    }), 
    { headers }
  );
  check(sendCoinRes, {
    'отправка монет успешна': (r) => r.status === 200,
  });

  // Тест покупки предмета
  const buyRes = http.get(`${BASE_URL}/api/buy/item1`, { headers });
  check(buyRes, {
    'покупка предмета успешна': (r) => r.status === 200,
  });

  sleep(1);
} 