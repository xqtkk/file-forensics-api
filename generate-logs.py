import json
import random
from datetime import datetime, timedelta

random.seed(42)

events = []
start_time = datetime(2024, 6, 15, 9, 0, 0)

# ============ НОРМАЛЬНАЯ АКТИВНОСТЬ ============
normal_users = ["alice", "bob", "carol", "dave"]
normal_ips = ["198.51.100.10", "198.51.100.11", "198.51.100.12"]
normal_actions = [
    "DescribeInstances", "ListBuckets", "GetObject",
    "PutObject", "DescribeSecurityGroups", "GetCallerIdentity"
]

current_time = start_time
for i in range(30):     # стало 30
    events.append({
        "eventTime": current_time.strftime("%Y-%m-%dT%H:%M:%SZ"),
        "eventName": random.choice(normal_actions),
        "eventSource": random.choice(["ec2.amazonaws.com", "s3.amazonaws.com", "sts.amazonaws.com"]),
        "awsRegion": random.choice(["us-east-1", "eu-west-1"]),
        "sourceIPAddress": random.choice(normal_ips),
        "userIdentity": {
            "type": "IAMUser",
            "userName": random.choice(normal_users)
        }
    })
    current_time += timedelta(seconds=random.randint(10, 120))

# ============ АТАКА ============
attack_start = start_time + timedelta(hours=2)
attacker_ip = "203.0.113.42"  # подозрительный IP
attacker_user = "compromised-user"

attack_chain = [
    ("GetCallerIdentity", "sts.amazonaws.com"),      # 1. Initial Access — проверка, кто мы
    ("ListBuckets", "s3.amazonaws.com"),             # 2. Discovery — что есть
    ("DescribeInstances", "ec2.amazonaws.com"),      # 2. Discovery — какие машины
    ("CreateUser", "iam.amazonaws.com"),             # 3. Privilege Escalation — создаём юзера
    ("CreateAccessKey", "iam.amazonaws.com"),        # 3. Privilege Escalation — ключ для юзера
    ("AttachUserPolicy", "iam.amazonaws.com"),       # 3. Privilege Escalation — даём админские права
    ("CreateUser", "iam.amazonaws.com"),             # 4. Persistence — ещё один юзер
    ("AttachUserPolicy", "iam.amazonaws.com"),       # 4. Persistence — снова админ
    ("ListBuckets", "s3.amazonaws.com"),             # 5. Exfiltration — смотрим бакеты
    ("GetObject", "s3.amazonaws.com"),               # 5. Exfiltration — качаем данные
    ("GetObject", "s3.amazonaws.com"),               # 5. Exfiltration
    ("GetObject", "s3.amazonaws.com"),               # 5. Exfiltration
    ("DeleteBucket", "s3.amazonaws.com"),            # 6. Impact — удаляем следы
]

attack_time = attack_start
for action, source in attack_chain:
    events.append({
        "eventTime": attack_time.strftime("%Y-%m-%dT%H:%M:%SZ"),
        "eventName": action,
        "eventSource": source,
        "awsRegion": "us-east-1",
        "sourceIPAddress": attacker_ip,
        "userIdentity": {
            "type": "IAMUser",
            "userName": attacker_user
        }
    })
    attack_time += timedelta(seconds=random.randint(5, 30))

# ============ ЕЩЁ НОРМАЛЬНАЯ АКТИВНОСТЬ (после атаки) ============
current_time = attack_time + timedelta(minutes=30)
for i in range(10):     # стало 10
    events.append({
        "eventTime": current_time.strftime("%Y-%m-%dT%H:%M:%SZ"),
        "eventName": random.choice(normal_actions),
        "eventSource": random.choice(["ec2.amazonaws.com", "s3.amazonaws.com", "sts.amazonaws.com"]),
        "awsRegion": random.choice(["us-east-1", "eu-west-1"]),
        "sourceIPAddress": random.choice(normal_ips),
        "userIdentity": {
            "type": "IAMUser",
            "userName": random.choice(normal_users)
        }
    })
    current_time += timedelta(seconds=random.randint(10, 120))

# Сортируем по времени (важно для timeline!)
events.sort(key=lambda e: e["eventTime"])

with open("attack-logs.json", "w", encoding="utf-8") as f:
    json.dump(events, f, ensure_ascii=False, indent=2)

print(f"Сгенерировано {len(events)} событий")
print(f"Нормальных: ~250, атакующих: {len(attack_chain)}")
print(f"Атакующий IP: {attacker_ip}")
print(f"Атакующий пользователь: {attacker_user}")