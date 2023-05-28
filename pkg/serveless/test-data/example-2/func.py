def main(params):
    x = params["x"]
    y = params["y"]
    xx = x + y
    yy = x - y
    resp = {
        "x": xx,
        "y": yy
    }
    return resp