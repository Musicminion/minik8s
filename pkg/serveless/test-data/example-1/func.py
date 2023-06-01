def main(params):
    x = params["x"]
    y = params["y"]
    xx = x - y
    yy = y - x
    resp = {
        "x": xx,
        "y": yy
    }
    return resp