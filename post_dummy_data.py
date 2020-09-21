#!/usr/bin/python
# post_dummy_data.py

import json
import random
import string
import urllib.request


API_SERVER = "http://localhost:8080"
API_USERS  = API_SERVER + "/users"
API_GROUPS = API_SERVER + "/groups"

# the API server is only expecting a non-empty auth token, if at all.
# NOTE: even though the Authorization header isn't required if the API server
#       is run with insecure_requests_mode=true, it's provided so that this
#       script works in either case
AUTH_TOKEN  = "dummy_token"
AUTH_HEADER = {"Authorization": "Bearer " + AUTH_TOKEN}


def generateID():
    letters = string.ascii_lowercase
    return ''.join(random.choice(letters) for i in range(10))


def jsonRequest(url="", method="", headers={}, jsonData={}):
    data = json.dumps(jsonData).encode("utf-8")
    req = urllib.request.Request(url=url, data=data, method=method)
    for h in headers:
        req.add_header(h, headers[h])

    try:
        resp = urllib.request.urlopen(req)
        if resp.status != 200:
            raise Exception("unexpected resp status", url, resp.status,
                    resp.reason)
    except Exception as e:
        raise e

    # https://docs.python.org/3/library/json.html
    body = resp.read()
    return json.loads(body.decode("utf-8"))


def createUser(first_name, last_name, groups=[]):
    resp = jsonRequest(url=API_USERS, method="POST", headers=AUTH_HEADER,
        jsonData={
            "first_name": first_name,
            "last_name": last_name,
            "userid": "user" + generateID(),
            "groups": groups,
        })
    return resp['user']


def createGroup(name):
    resp = jsonRequest(url=API_GROUPS, method="POST", headers=AUTH_HEADER,
        jsonData={
            "name": "group" + generateID(),
        })
    return resp['group']


def setUserMembership(user_id, group_names):
    url = API_USERS + "/" + user_id
    resp = jsonRequest(url=url, method="PUT", headers=AUTH_HEADER, jsonData={
        "groups": group_names,
    })
    return resp


def setGroupMembership(group_name, user_ids):
    url = API_GROUPS + "/" + group_name
    resp = jsonRequest(url=url, method="PUT", headers=AUTH_HEADER, jsonData={
        "userids": user_ids,
    })
    return resp


def run():
    user_ids = []
    for i in range(10):
        user = createUser("user" + str(i), "last" + str(i))
        user_ids.append(user['userid'])

    group_names = []
    for i in range(5):
        group = createGroup("group" + str(i))
        group_names.append(group['name'])

    setGroupMembership(group_names[0], user_ids[0:3])
    setGroupMembership(group_names[1], user_ids[1:3])
    setGroupMembership(group_names[2], user_ids[2:3])
    setGroupMembership(group_names[3], user_ids[3:8])
    setGroupMembership(group_names[4], user_ids[4:9])

    user = createUser("allgroup", "every_single_group", group_names)
    user_ids.append(user['userid'])


if __name__ == "__main__":
    run()
