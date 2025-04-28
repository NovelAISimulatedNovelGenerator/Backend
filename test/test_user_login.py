# -*- coding: utf-8 -*-
"""
test_user_login.py
用户登录接口（/api/user/login）API单元测试

- 覆盖正常登录、密码错误、用户不存在、参数缺失等情况
- 需先启动后端服务
- 依赖 pytest fixture auto_register_and_login
"""
import requests
import pytest
from conftest import TEST_USERNAME, TEST_PASSWORD

BASE_URL = "http://127.0.0.1:8888"
LOGIN_URL = f"{BASE_URL}/api/user/login"

class TestUserLogin:
    """
    /api/user/login 用户登录接口测试
    """
    def test_login_success(self, auto_register_and_login):
        """
        正常登录，断言返回 token
        """
        resp = requests.post(LOGIN_URL, json={"username": TEST_USERNAME, "password": TEST_PASSWORD})
        assert resp.status_code == 200
        data = resp.json()
        assert "token" in data
        assert data["token"]

    def test_login_wrong_password(self, auto_register_and_login):
        """
        密码错误，登录失败
        """
        resp = requests.post(LOGIN_URL, json={"username": TEST_USERNAME, "password": "wrongpass"})
        assert resp.status_code == 200 or resp.status_code == 401
        data = resp.json()
        assert data["code"] != 200
        # 兼容英文与中文错误信息
        assert (
            "错误" in data["message"] or
            "失败" in data["message"] or
            "incorrect" in data["message"].lower() or
            "not exist" in data["message"].lower() or
            "error" in data["message"].lower()
        )

    def test_login_user_not_exist(self):
        """
        用户不存在，登录失败
        """
        resp = requests.post(LOGIN_URL, json={"username": "not_exist_user_abc", "password": "any"})
        assert resp.status_code == 200 or resp.status_code == 401
        data = resp.json()
        assert data["code"] != 200
        # 兼容英文与中文错误信息
        assert (
            "不存在" in data["message"] or
            "失败" in data["message"] or
            "incorrect" in data["message"].lower() or
            "not exist" in data["message"].lower() or
            "error" in data["message"].lower()
        )

    def test_login_missing_params(self):
        """
        缺少参数，登录失败
        """
        resp = requests.post(LOGIN_URL, json={"username": TEST_USERNAME})  # 缺 password
        assert resp.status_code in (400, 401, 200)
        data = resp.json()
        assert data["code"] in (400, 401, 500)

# 运行方法：pytest test_user_login.py
