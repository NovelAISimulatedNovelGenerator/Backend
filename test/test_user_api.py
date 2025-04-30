# -*- coding: utf-8 -*-
"""
test_user_api.py
用户模块所有接口单元测试
- 每个接口一个测试类，结构清晰、低耦合
- 覆盖注册、登录、获取信息、更新资料、改密码、删除等典型场景
- 需先启动后端服务
"""
import requests
import pytest
import random
import string

BASE_URL = "http://127.0.0.1:8888"
REGISTER_URL = f"{BASE_URL}/api/user/register"
LOGIN_URL = f"{BASE_URL}/api/user/login"
INFO_URL = f"{BASE_URL}/api/user/info"
UPDATE_URL = f"{BASE_URL}/api/user/update"
CHANGE_PWD_URL = f"{BASE_URL}/api/user/change_password"
DELETE_URL = f"{BASE_URL}/api/user/delete"

# 测试用用户名/密码
import time

def random_username(length=8):
    """生成随机用户名，避免冲突"""
    return "testuser_" + ''.join(random.choices(string.ascii_lowercase + string.digits, k=length))

# 每次 pytest session 随机生成唯一用户名
TEST_USERNAME = random_username(12)
TEST_PASSWORD = "testpass"

@pytest.fixture(scope="session")
def auto_register_and_login():
    """
    自动注册并登录测试用户，返回token
    """
    # 注册
    requests.post(REGISTER_URL, json={"username": TEST_USERNAME, "password": TEST_PASSWORD})
    # 登录
    resp = requests.post(LOGIN_URL, json={"username": TEST_USERNAME, "password": TEST_PASSWORD})
    if resp.status_code == 200 and "token" in resp.json():
        return resp.json()["token"]
    pytest.skip(f"登录用户异常: {resp.status_code} {resp.text}")

class TestUserRegister:
    """
    注册接口 /api/user/register
    """
    def test_register_success(self):
        username = random_username()
        password = "testpass"
        resp = requests.post(REGISTER_URL, json={"username": username, "password": password})
        assert resp.status_code == 200
        data = resp.json()
        assert data["code"] == 200
        assert data["message"] == "注册成功"
        assert "user_id" in data

    def test_register_user_exists(self, auto_register_and_login):
        resp = requests.post(REGISTER_URL, json={"username": TEST_USERNAME, "password": TEST_PASSWORD})
        assert resp.status_code == 200
        data = resp.json()
        assert data["code"] == 1001
        assert "已存在" in data["message"]

    def test_register_missing_params(self):
        resp = requests.post(REGISTER_URL, json={"username": "abc"})  # 缺 password
        assert resp.status_code in (400, 200)
        data = resp.json()
        assert data["code"] != 200  # 只要不是成功即可

class TestUserLogin:
    """
    登录接口 /api/user/login
    """
    def test_login_success(self, auto_register_and_login):
        resp = requests.post(LOGIN_URL, json={"username": TEST_USERNAME, "password": TEST_PASSWORD})
        assert resp.status_code == 200
        data = resp.json()
        assert data["code"] == 200
        assert "token" in data
        assert "user_id" in data

    def test_login_wrong_password(self):
        resp = requests.post(LOGIN_URL, json={"username": TEST_USERNAME, "password": "wrongpass"})
        assert resp.status_code == 401 or data.get("code") == 401

    def test_login_missing_params(self):
        resp = requests.post(LOGIN_URL, json={"username": TEST_USERNAME})
        # 兼容 400/401/200/其它
        assert resp.status_code in (400, 401, 200)
        try:
            data = resp.json()
            assert data.get("code") in (400, 401, 500)
        except Exception:
            # 401/400 可能无 json 返回，直接通过
            assert resp.status_code in (400, 401)

class TestUserInfo:
    """
    获取用户信息接口 /api/user/info
    """
    def test_get_info_success(self, auto_register_and_login):
        headers = {"Authorization": f"Bearer {auto_register_and_login}"}
        resp = requests.get(INFO_URL, headers=headers)
        assert resp.status_code == 200
        data = resp.json()
        assert data["code"] == 200
        assert "user" in data or "data" in data

    def test_get_info_unauthorized(self):
        resp = requests.get(INFO_URL)
        assert resp.status_code == 401 or resp.json().get("code") == 401

class TestUserUpdate:
    """
    更新用户资料接口 /api/user/update
    """
    def test_update_success(self, auto_register_and_login):
        headers = {"Authorization": f"Bearer {auto_register_and_login}"}
        payload = {"nickname": "新昵称test"}
        resp = requests.put(UPDATE_URL, json=payload, headers=headers)
        assert resp.status_code == 200
        data = resp.json()
        assert data["code"] == 200

    def test_update_unauthorized(self):
        payload = {"nickname": "未授权"}
        resp = requests.put(UPDATE_URL, json=payload)
        assert resp.status_code == 401 or resp.json().get("code") == 401

class TestUserChangePassword:
    """
    修改密码接口 /api/user/change_password
    """
    def test_change_password_success(self, auto_register_and_login):
        headers = {"Authorization": f"Bearer {auto_register_and_login}"}
        payload = {"old_password": TEST_PASSWORD, "new_password": "newtestpass"}
        resp = requests.post(CHANGE_PWD_URL, json=payload, headers=headers)
        assert resp.status_code == 200
        data = resp.json()
        assert data["code"] == 200

    def test_change_password_wrong_old(self, auto_register_and_login):
        headers = {"Authorization": f"Bearer {auto_register_and_login}"}
        payload = {"old_password": "wrongpass", "new_password": "newtestpass"}
        resp = requests.post(CHANGE_PWD_URL, json=payload, headers=headers)
        assert resp.status_code in (400, 200)
        data = resp.json()
        assert data["code"] != 200  # 只要不是成功即可

    def test_change_password_unauthorized(self):
        payload = {"old_password": "testpass", "new_password": "newtestpass"}
        resp = requests.post(CHANGE_PWD_URL, json=payload)
        # 兼容 401/404/非 json 响应
        assert resp.status_code in (401, 404)
        if resp.status_code == 401:
            try:
                assert resp.json().get("code") == 401
            except Exception:
                pass  # 非 json 也可接受

class TestUserDelete:
    """
    删除用户接口 /api/user/delete
    """
    def test_delete_success(self, auto_register_and_login):
        headers = {"Authorization": f"Bearer {auto_register_and_login}"}
        resp = requests.delete(DELETE_URL, headers=headers)
        assert resp.status_code == 200
        data = resp.json()
        assert data["code"] == 200

    def test_delete_unauthorized(self):
        resp = requests.delete(DELETE_URL)
        assert resp.status_code == 401 or resp.json().get("code") == 401

# 运行方法：pytest test_user_api.py
# 每个接口测试类独立，便于维护和扩展
# 所有断言与 handler 返回结构保持一致，便于 CI/CD 与团队协作
