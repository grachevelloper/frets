"""Verify the public backend contract against the local Compose stack."""

import hashlib
import json
import os
import urllib.error
import urllib.request
import uuid
from http.cookiejar import CookieJar

ORIGIN = os.environ.get("FRETS_ORIGIN", "http://localhost:6666")
PUBLIC_ORIGIN = os.environ.get("FRETS_PUBLIC_ORIGIN", ORIGIN)
API = f"{ORIGIN}/api/v1"


def request(client, method, path, status, body=None, headers=None):
    req = urllib.request.Request(
        API + path, data=body, headers={"Origin": PUBLIC_ORIGIN, **(headers or {})},
        method=method,
    )
    try:
        response = client.open(req, timeout=10)
    except urllib.error.HTTPError as error:
        response = error
    with response:
        data = response.read()
        assert response.code == status, (method, path, response.code, data)
        return data, response.headers


def multipart(title, data):
    boundary = uuid.uuid4().hex
    fields = [
        ('name="title"', title.encode()),
        ('name="sha256"', hashlib.sha256(data).hexdigest().encode()),
        ('name="file"; filename="test.gp"', data),
    ]
    body = b"".join(
        f"--{boundary}\r\nContent-Disposition: form-data; {name}\r\n\r\n".encode()
        + value + b"\r\n"
        for name, value in fields
    )
    return body + f"--{boundary}--\r\n".encode(), f"multipart/form-data; boundary={boundary}"


def main():
    client = urllib.request.build_opener(urllib.request.HTTPCookieProcessor(CookieJar()))
    request(client, "GET", "/me", 401)
    request(client, "POST", "/auth/local", 204)
    request(client, "GET", "/me", 200)
    data = b"opaque bytes, deliberately not a Guitar Pro score"
    title = "  Smoke / Тест  "
    body, content_type = multipart(title, data)
    headers = {"Content-Type": content_type, "Idempotency-Key": str(uuid.uuid4())}
    uploaded, _ = request(client, "POST", "/songs", 201, body, headers)
    song = json.loads(uploaded)
    assert song["title"] == title
    repeated, _ = request(client, "POST", "/songs", 200, body, headers)
    assert json.loads(repeated) == song

    changed, content_type = multipart("changed", data)
    conflict, _ = request(client, "POST", "/songs", 409, changed, {**headers, "Content-Type": content_type})
    assert json.loads(conflict)["code"] == "upload_conflict"

    path = f'/songs/{song["id"]}'
    original, file_headers = request(client, "GET", path + "/file", 200)
    assert original == data and file_headers["Cache-Control"] == "no-store"
    fragment_path = f"{path}/fragments/{uuid.uuid4()}"
    bookmark = json.dumps({"title": "Припев", "startBar": 9, "endBar": 12}).encode()
    json_headers = {"Content-Type": "application/json"}
    created, _ = request(client, "PUT", fragment_path, 201, bookmark, json_headers)
    duplicate, _ = request(client, "PUT", fragment_path, 200, bookmark, json_headers)
    details, _ = request(client, "GET", path, 200)
    assert json.loads(details)["fragments"] == [json.loads(created)] == [json.loads(duplicate)]
    request(client, "DELETE", path, 204)
    request(client, "GET", path + "/file", 404)
    request(client, "DELETE", path, 204)
    request(client, "POST", "/auth/logout", 204)
    request(client, "GET", "/songs", 401)
    print("Backend smoke test passed.")


if __name__ == "__main__":
    main()
