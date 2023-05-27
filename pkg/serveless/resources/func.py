def main(params):
    x = params["x"]
    y = params["y"]
    x = x + y
    resp = {
        "sum": x
    }
    return resp