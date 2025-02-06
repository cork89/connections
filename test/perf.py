from time import perf_counter

import requests

final_results = []

for i in range(10000):
    home_start = perf_counter()
    home_res = requests.get("http://localhost:8000")
    home_end = perf_counter()
    create_res = requests.get("http://localhost:8000/create/")
    create_end = perf_counter()
    game_res = requests.get("http://localhost:8000/game/abc123/")
    game_end = perf_counter()

    home_time = home_end - home_start
    create_time = create_end - home_end
    game_time = game_end - create_end

    final_results.append(
        "{},{},{:.2f},{},{:.2f},{},{:.2f}\n".format(
            i,
            home_res.status_code,
            home_time * 1000,
            create_res.status_code,
            create_time * 1000,
            game_res.status_code,
            game_time * 1000,
        )
    )

with open("results.csv", "w") as file:
    file.write(
        "cnt,home_status_code,home_time,create_status_code,create_time,game_status_code,game_time\n"
    )
    file.writelines(final_results)
