import brewer2mpl
import tufte
import matplotlib
import matplotlib.pyplot as plt

matplotlib.use("pgf")

golden_ratio = (5**.5 - 1) / 2

pgf_with_pdflatex = {
#    "pgf.texsystem": "pdflatex",
#    "pgf.rcfonts": False,
#    "pgf.preamble": r"""
#    \usepackage[T1]{fontenc}
#    \usepackage{newtxmath}
#    \usepackage{newtxtext}
#    \usepackage{microtype}
#    \usepackage{pifont}
#         """,
#    "text.usetex": True,
#    "font.family": "serif",
    "figure.figsize": [3.336,3.15*golden_ratio],  
    "font.sans-serif": [],
    "font.serif": [],
    "font.monospace": [],
    "axes.labelsize": 8, 
    "font.size": 9,
    "legend.fontsize": 8, 
    "xtick.labelsize": 8,
    "ytick.labelsize": 8,
    "lines.markersize": 3, 
    "lines.markeredgewidth": 0,
    "axes.linewidth": 0.5,
}


matplotlib.rcParams.update(pgf_with_pdflatex)

import matplotlib.style
import matplotlib.pyplot as plt
from matplotlib.lines import Line2D

_markers = ["o", "v", "s", "*", "D", "^"]
hash_markers = _markers
mix_markers = _markers

def as_number(x, pos):
    return "%d" % x

def as_number_m(x, pos):
    return "%d" % (x/(2**20))

def millions(x, pos):
    return "%d" % (x/1000000)

def megabytes10(x, pos):
  """Formatter for Y axis, values are in megabytes"""
  if x < 1000:
      return '%d B' % (x)
  elif x < 1000 * 1000:
      return '%1.0f kB' % (x/1000)
  else:
      return '%1.0f mB' % (x/(1000*1000))

def megabytes(x, pos):
  """Formatter for Y axis, values are in megabytes"""
  if x < 1024:
      return '%d B' % (x)
  elif x < 1024 * 1024:
      return '%1.0f KiB' % (x/1024)
  else:
      return '%1.0f MiB' % (x/(1024*1024))

def bits(x, pos):
  """Formatter for Y axis, values are in megabytes"""
  if x < 8:
      return '%d b' % (x)
  x /= 8
  if x < 1024:
      return '%d B' % (x)
  elif x < 1024 * 1024:
      return '%1.0f KiB' % (x/1024)
  else:
      return '%1.0f MiB' % (x/(1024*1024))

# brewer2mpl.get_map args: set name  set type  number of colors
bmap1 = brewer2mpl.get_map('Set1', 'Qualitative', 7)
bmap2 = brewer2mpl.get_map('Dark2', 'Qualitative', 7)
hash_colors = bmap1.mpl_colors
mix_colors = bmap2.mpl_colors
fig, ax = plt.subplots()
tufte.tuftestyle(ax)
plt.tight_layout()
plt.grid(axis='y', color="0.9", linestyle='-', linewidth=1)

def save_fig(fig, out_name, size=None, pad = 0, width=3.336, tight=True):
    if size == None:
        size = [width, width*golden_ratio]
    fig.set_size_inches(size)
    if tight:
        fig.tight_layout()
    plt.savefig(out_name, dpi=600, bbox_inches='tight', pad_inches = pad)
  
def setup_columns(f):
    return f.readline().split()

def col(pieces, cols, name):
    return pieces[cols.index(name)]

def remove_chart_junk(plt, ax, grid=False, ticks=False):
    ax.spines['top'].set_visible(False)
    ax.spines['right'].set_visible(False)
    ax.get_xaxis().tick_bottom()
    ax.get_yaxis().tick_left()
    #ax.xaxis.set_ticks_position('none')
    #if not ticks:
    #    ax.yaxis.set_ticks_position('none')
    #else:
    #    plt.minorticks_off()

    ax.set_axisbelow(True)
    if grid:
        #plt.grid(b=True, which='major', color='0.9', linestyle='-')
    #else:
        ax.yaxis.grid(which='major', color='0.9', linestyle='--')


def reformat_large_tick_values(tick_val, pos):
    """
    Turns large tick values (in the billions, millions and thousands) such as 4500 into 4.5K and also appropriately turns 4000 into 4K (no zero after the decimal).
    """
    if tick_val >= 1000000000:
        val = round(tick_val/1000000000, 1)
        new_tick_format = '{:}B'.format(val)
    elif tick_val >= 1000000:
        val = round(tick_val/1000000, 1)
        new_tick_format = '{:}M'.format(val)
    elif tick_val >= 1000:
        val = round(tick_val/1000, 1)
        new_tick_format = '{:}K'.format(val)
    elif tick_val < 1000:
        new_tick_format = round(tick_val, 1)
    else:
        new_tick_format = tick_val

    # make new_tick_format into a string value
    new_tick_format = str(new_tick_format)
    
    # code below will keep 4.5M as is but change values such as 4.0M to 4M since that zero after the decimal isn't needed
    index_of_decimal = new_tick_format.find(".")
    
    if index_of_decimal != -1:
        value_after_decimal = new_tick_format[index_of_decimal+1]
        if value_after_decimal == "0":
            # remove the 0 after the decimal point since it's not needed
            new_tick_format = new_tick_format[0:index_of_decimal] + new_tick_format[index_of_decimal+2:]
            
    return new_tick_format
