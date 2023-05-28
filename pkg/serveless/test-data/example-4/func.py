def main(params):
    x = params["x"]
    y = params["y"]
    xx = y + x * x
    yy = x + y * y
    resp = {
        "x": xx,
        "y": yy
    }
    return resp