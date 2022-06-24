package scheduler


type Signals struct {
	isPileReady     * chan bool	//
	stopPile         chan bool	//
}