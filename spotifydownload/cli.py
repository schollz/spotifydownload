# -*- coding: utf-8 -*-

"""Console script for spotifydownload."""
import sys
import click


@click.command()
@click.argument('bearer')
@click.argument('playlistid')
def main(bearer,playlistid):
    """Console script for spotifydownload."""
    click.echo("Replace this message by putting your code into "
               "spotifydownload.cli.main")
    click.echo("See click documentation at http://click.pocoo.org/")
    return 0


if __name__ == "__main__":
    sys.exit(main())  # pragma: no cover
