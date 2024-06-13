import "@stdlib/ownable";
import "./oracleProtocol.tact";

message (0x1af44a9f) MoveTo {
    newAddress: Address;
    moveCompleted: Bool;
}

trait  Movable with Ownable {

    owner: Address;
    newAddress: Address;
    moved: Bool;

    receive (msg: MoveTo) {
        self.requireOwner();
        self.newAddress = msg.newAddress;
        self.moved = msg.moveCompleted;
        let sb: StringBuilder = beginString();
        sb.append("Moved to: ");
        sb.append(self.newAddress.toString());
        emit(sb.toString().asComment());
    }

    receive ("NewAddress") {
        self.replyNewAddress(0);
    }

    get fun newAddress() : Address {
        return self.newAddress;
    }

    fun replyNewAddress(queryID: Int) {
        self.reply(OracleNewAddressResponse{
            newAddress: self.newAddress,
            queryID: queryID,
        }.toCell());
    }

    fun requireNotMoved() {
        require(!self.moved, "Moved to new address");
    }

}
