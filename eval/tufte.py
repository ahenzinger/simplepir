
import matplotlib 
def tuftestyle(ax):
    """Styles an axes object to have minimalist graph features"""
    #ax.grid(True, 'major', color='w', linestyle='-', linewidth=1.5,alpha=0.6)
    #ax.patch.set_facecolor('white')
    
    ax.set_axisbelow(True)
    
    ax.spines["right"].set_visible(False)
    ax.spines["top"].set_visible(False)
    ax.spines["bottom"].set_visible(False)
    ax.spines["left"].set_visible(False)
    #ax.yaxis.set_major_locator(MultipleLocator( (ax.get_yticks()[-1]-ax.get_yticks()[0]) / 0.1 ))
    #ax.get_xaxis().tick_bottom()
    #ax.get_yaxis().tick_left()
      
    #restyle the tick lines
    for line in ax.get_xticklines() + ax.get_yticklines():
        line.set_markersize(5)
        line.set_color("black")
        line.set_markeredgewidth(0)
    
    for line in ax.xaxis.get_ticklines(minor=True) + ax.yaxis.get_ticklines(minor=True):
        line.set_markersize(0)
    
    matplotlib.rcParams['xtick.direction'] = 'out'
    matplotlib.rcParams['ytick.direction'] = 'in'
    ax.xaxis.set_ticks_position('bottom')
    ax.yaxis.set_ticks_position('left')

