from argparse import ArgumentParser
from metrics import FileMetrics
import logging


def main():
    parser = ArgumentParser(prog="capy", description="Code analyser for python")
    parser.add_argument(
        "-f",
        dest="filename",
        type=str,
        required=True,
        help="Python source file to analyse",
    )
    parser.add_argument(
        "--log",
        dest="loglevel",
        default="INFO",
        type=str,
        help="The log level to set, one of DEBUG, INFO, WARNING, ERROR",
    )
    args = parser.parse_args()
    # set the log level
    log_level = getattr(logging, args.loglevel.upper(), None)
    if not isinstance(log_level, int):
        raise ValueError("Invalid log level: %s" % args.loglevel)
    logging.basicConfig(
        level=log_level,
        format="%(asctime)s, %(levelname)s, %(module)s/%(funcName)s:%(lineno)d: %(message)s",
    )

    fms = FileMetrics(args.filename)
    fms.generate_metrics()

    print_metrics(fms)


def print_metrics(fms: FileMetrics):
    print(f"File metrics: {fms}")


if __name__ == "__main__":
    main()
