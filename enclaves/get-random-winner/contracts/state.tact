const stateInit: Int = 0;
const stateRock: Int = 1;
const stateRoll: Int = 2;
const stateWaitReveal: Int = 3;
const stateCompleted: Int = 15;

const stakeMinValue: Int = ton("0.02");
const waitRevealMaxPeriod: Int = 120;

struct EventState {
    state: Int as uint8;
    changed: Int as uint8;
    updated: Int as uint32;
    completionTime: Int as uint64;
    stakeOnRock: Int as uint64;
    stakeOnRoll: Int as uint64;
    doraId: Int as uint64;
    name: String;
    prize: Int as uint64;
}

fun initialEventState(completionTime: Int) : EventState {
    return EventState{
        state: stateInit,
        changed: stateInit,
        updated: now(),
        completionTime: completionTime,
        stakeOnRock: 0,
        stakeOnRoll: 0,
        doraId: 0,
        name: "No winner yet",
        prize: 0,
    };
}

extends mutates fun rock(self: EventState, value: Int) : String {
    if (now() > self.completionTime) {
        self.state = stateCompleted;
        return "Event is completed.";
    }
    if (value < stakeMinValue) { return "Value amount is not enough to Rock"; }
    if (self.state != stateRock) { return "Rejected. We're rolling to next BUIDl."; }
    self.stake(value, 0);
    self.update(stateRock);
    return "";
}

extends mutates fun roll(self: EventState, value: Int) : String {
    if (now() > self.completionTime) {
        self.state = stateCompleted;
        return "Event is completed.";
    }
    let timeout : Bool = now() - self.updated > waitRevealMaxPeriod;
    if (value < stakeMinValue) { return "Value amount is not enough to Roll"; }
    if (self.state != stateRock && self.state != stateInit && !timeout) { return "Rejected. We're already rolling."; }
    self.stake(0, value);

    if ((self.state == stateRoll || self.state == stateWaitReveal) && timeout) {
        // reset the state, to perform the reroll
        self.update(stateInit);
    }

    let nextState: Int = stateRoll;
    if (self.state == stateRock && self.stakeOnRoll <= self.stakeOnRock) {
        nextState = stateRock;
    }
    self.update(nextState);
    return "";
}

extends mutates fun waitReveal(self: EventState) : String {
    if (now() > self.completionTime) {
        self.state = stateCompleted;
        return "Event is completed.";
    }
    if (self.state != stateRoll) { return "Rejected. Not in a Roll state."; }
    self.update(stateWaitReveal);
    return "";
}

extends mutates fun reveal(self: EventState, doraId: Int, name: String) : String {
    if (now() > self.completionTime) {
        self.state = stateCompleted;
        return "Event is completed.";
    }
    if (self.state != stateWaitReveal) { return "Rejected. Not in WaitReveal state."; }
    self.doraId = doraId;
    self.name = name;
    self.stakeOnRock = stakeMinValue;
    self.stakeOnRoll = 0;
    self.update(stateRock);
    return "";
}

extends mutates fun stake(self: EventState, rock: Int, roll: Int) {
    self.stakeOnRock += rock;
    self.stakeOnRoll += roll;
}


extends mutates fun update(self: EventState, state: Int){
    self.updated = now();
    self.changed = (self.state != state) ? state : stateInit;
    self.state = state;
}

extends fun requireCompleted(self: EventState) {
    require(self.state == stateCompleted, "Event is not completed.");
}

extends fun requireNotOutdated(self: EventState, timestamp: Int) {
    require(timestamp >= self.updated, "Timestamp is outdated.");
}
